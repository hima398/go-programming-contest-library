package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	inFlag := flag.String("in", "", "入力 main.go のパス（必須）")
	outFlag := flag.String("out", "", "出力先パス（省略時は stdout）")
	flag.Parse()

	if *inFlag == "" {
		fmt.Fprintln(os.Stderr, "usage: ac-bundle -in <main.go> [-out <submit.go>]")
		os.Exit(1)
	}

	result, err := bundle(*inFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *outFlag == "" {
		fmt.Print(string(result))
	} else {
		if err := os.WriteFile(*outFlag, result, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}
}

type pkgMeta struct {
	importPath string
	shortName  string
	dir        string
}

type bundler struct {
	fset *token.FileSet
	// pkgRename: pkgShortName → {originalName → resolvedName}
	pkgRename map[string]map[string]string
}

func bundle(mainPath string) ([]byte, error) {
	absMain, err := filepath.Abs(mainPath)
	if err != nil {
		return nil, err
	}

	modRoot, modName, err := findModule(filepath.Dir(absMain))
	if err != nil {
		return nil, err
	}

	// go.mod の require / replace ブロックを解析
	gomodData, err := os.ReadFile(filepath.Join(modRoot, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("go.mod 読み込みエラー: %w", err)
	}
	gomodStr := string(gomodData)
	requires := parseRequires(gomodStr)
	replaces := parseReplaces(gomodStr, modRoot)

	fset := token.NewFileSet()
	mainFile, err := parser.ParseFile(fset, absMain, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("main.go parse error: %w", err)
	}

	// import を自モジュール / 外部モジュール / 標準ライブラリに分類
	var localPkgs []pkgMeta
	inlinedPaths := map[string]bool{} // インライン展開するパス

	for _, imp := range mainFile.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		shortName := path[strings.LastIndex(path, "/")+1:]

		if strings.HasPrefix(path, modName+"/") {
			// 自モジュール: modRoot 以下で解決
			rel := strings.TrimPrefix(path, modName+"/")
			localPkgs = append(localPkgs, pkgMeta{
				importPath: path,
				shortName:  shortName,
				dir:        filepath.Join(modRoot, rel),
			})
			inlinedPaths[path] = true
		} else {
			// 外部モジュール: replace → GOMODCACHE の順で解決
			extMod, ver, subPath, ok := findModuleForImport(path, requires)
			if !ok {
				// require になくても replace だけある場合（go.sum 管理外のローカル置換）
				extMod, _, subPath, ok = findModuleForImport(path, replaces)
				if !ok {
					continue
				}
				ver = ""
			}
			var srcDir string
			if localPath, ok := replaces[extMod]; ok {
				// replace ディレクティブ優先（ローカルチェックアウト）
				srcDir = localPath
			} else {
				// モジュールキャッシュから解決
				srcDir, err = resolveModuleDir(extMod, ver)
				if err != nil {
					return nil, fmt.Errorf("外部モジュール解決失敗 %s: %w", path, err)
				}
			}
			pkgDir := filepath.Join(srcDir, subPath)
			localPkgs = append(localPkgs, pkgMeta{
				importPath: path,
				shortName:  shortName,
				dir:        pkgDir,
			})
			inlinedPaths[path] = true
		}
	}

	if len(localPkgs) == 0 {
		var buf strings.Builder
		if err := format.Node(&buf, fset, mainFile); err != nil {
			return nil, err
		}
		return []byte(buf.String()), nil
	}

	// 各パッケージのトップレベル名と宣言を収集
	type pkgInfo struct {
		meta    pkgMeta
		decls   []ast.Decl
		imports []string
		names   []string
	}
	allPkgs := []pkgInfo{}
	nameCount := map[string]int{}

	for _, pkg := range localPkgs {
		decls, imports, names, err := parsePackage(fset, pkg.dir)
		if err != nil {
			return nil, fmt.Errorf("package %s: %w", pkg.importPath, err)
		}
		allPkgs = append(allPkgs, pkgInfo{pkg, decls, imports, names})
		for _, n := range names {
			nameCount[n]++
		}
	}

	// rename マップを構築: pkgShortName → {oldName → newName}
	pkgRename := map[string]map[string]string{}
	for _, pi := range allPkgs {
		renames := map[string]string{}
		for _, name := range pi.names {
			if nameCount[name] > 1 {
				renames[name] = pi.meta.shortName + upperFirst(name)
			} else {
				renames[name] = name
			}
		}
		pkgRename[pi.meta.shortName] = renames
	}

	b := &bundler{fset: fset, pkgRename: pkgRename}

	// 標準 import を収集（インライン展開するパスは除去）
	stdImports := map[string]bool{}
	for _, imp := range mainFile.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !inlinedPaths[path] {
			stdImports[path] = true
		}
	}
	for _, pi := range allPkgs {
		for _, imp := range pi.imports {
			stdImports[imp] = true
		}
	}

	// main.go の AST を書き換え（pkg.Name → 解決済み名）
	b.rewriteFile(mainFile)

	// import 宣言を再構築
	rebuildImports(mainFile, stdImports)

	// ライブラリ宣言をリネームして末尾に追加
	for _, pi := range allPkgs {
		renamed := b.renameDecls(pi.decls, pi.meta.shortName)
		mainFile.Decls = append(mainFile.Decls, renamed...)
	}

	var buf strings.Builder
	if err := format.Node(&buf, fset, mainFile); err != nil {
		return nil, fmt.Errorf("format error: %w", err)
	}
	return []byte(buf.String()), nil
}

// findModuleForImport は importPath がどの外部モジュールに属するかを requires から探す。
// 最長一致（より具体的なモジュール名を優先）で返す。
func findModuleForImport(importPath string, requires map[string]string) (modName, version, subPath string, ok bool) {
	bestLen := 0
	for mod, ver := range requires {
		if importPath == mod || strings.HasPrefix(importPath, mod+"/") {
			if len(mod) > bestLen {
				bestLen = len(mod)
				modName, version = mod, ver
				subPath = strings.TrimPrefix(strings.TrimPrefix(importPath, mod), "/")
				ok = true
			}
		}
	}
	return
}

// resolveModuleDir は Go モジュールキャッシュからモジュールのローカルパスを返す。
func resolveModuleDir(modName, version string) (string, error) {
	cacheDir := os.Getenv("GOMODCACHE")
	if cacheDir == "" {
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("GOMODCACHE/GOPATH 未設定かつホームディレクトリ取得失敗: %w", err)
			}
			goPath = filepath.Join(home, "go")
		}
		cacheDir = filepath.Join(goPath, "pkg", "mod")
	}
	dir := filepath.Join(cacheDir, modName+"@"+version)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("モジュールキャッシュに %s@%s が見つかりません（%s）", modName, version, dir)
	}
	return dir, nil
}

// parseRequires は go.mod テキストから require ブロックを解析して {moduleName → version} を返す。
func parseRequires(gomod string) map[string]string {
	result := map[string]string{}
	lines := strings.Split(gomod, "\n")
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "require (" {
			inBlock = true
			continue
		}
		if inBlock {
			if trimmed == ")" {
				inBlock = false
				continue
			}
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 && !strings.HasPrefix(parts[0], "//") {
				result[parts[0]] = parts[1]
			}
		} else if strings.HasPrefix(trimmed, "require ") {
			rest := strings.TrimPrefix(trimmed, "require ")
			parts := strings.Fields(rest)
			if len(parts) >= 2 {
				result[parts[0]] = parts[1]
			}
		}
	}
	return result
}

// parseReplaces は go.mod の replace ディレクティブから {moduleName → localPath} を返す。
// ローカルパス（/ または . 始まり）の replace のみを対象とする。
func parseReplaces(gomod, modRoot string) map[string]string {
	result := map[string]string{}
	lines := strings.Split(gomod, "\n")
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "replace (" {
			inBlock = true
			continue
		}
		if inBlock && trimmed == ")" {
			inBlock = false
			continue
		}
		var replLine string
		if inBlock {
			replLine = trimmed
		} else if strings.HasPrefix(trimmed, "replace ") {
			replLine = strings.TrimPrefix(trimmed, "replace ")
		} else {
			continue
		}
		// "A [version] => B [version]"
		halves := strings.SplitN(replLine, " => ", 2)
		if len(halves) != 2 {
			continue
		}
		lhsParts := strings.Fields(halves[0])
		if len(lhsParts) == 0 {
			continue
		}
		modName := lhsParts[0]
		rhsParts := strings.Fields(strings.TrimSpace(halves[1]))
		if len(rhsParts) == 0 {
			continue
		}
		localPath := rhsParts[0]
		if !strings.HasPrefix(localPath, "/") && !strings.HasPrefix(localPath, ".") {
			continue // モジュール+バージョンによる置換はスキップ
		}
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(modRoot, localPath)
		}
		result[modName] = localPath
	}
	return result
}

// parsePackage はパッケージディレクトリの .go ファイルから宣言・imports・名前を収集する。
func parsePackage(fset *token.FileSet, pkgDir string) (decls []ast.Decl, imports []string, names []string, err error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, nil, nil, err
	}

	importSet := map[string]bool{}
	nameSet := map[string]bool{}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fpath := filepath.Join(pkgDir, name)
		f, err := parser.ParseFile(fset, fpath, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %w", fpath, err)
		}
		for _, imp := range f.Imports {
			importSet[strings.Trim(imp.Path.Value, `"`)] = true
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok == token.IMPORT {
					continue
				}
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						nameSet[s.Name.Name] = true
					case *ast.ValueSpec:
						for _, n := range s.Names {
							nameSet[n.Name] = true
						}
					}
				}
				decls = append(decls, d)
			case *ast.FuncDecl:
				if d.Recv == nil {
					nameSet[d.Name.Name] = true
				}
				decls = append(decls, d)
			}
		}
	}

	for imp := range importSet {
		imports = append(imports, imp)
	}
	for n := range nameSet {
		names = append(names, n)
	}
	return decls, imports, names, nil
}

// renameDecls はライブラリ宣言の衝突名をリネームし、内部参照も修正する。
func (b *bundler) renameDecls(decls []ast.Decl, pkgShortName string) []ast.Decl {
	renames := b.pkgRename[pkgShortName]
	for _, decl := range decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				if newName, ok := renames[d.Name.Name]; ok {
					d.Name.Name = newName
				}
			}
			b.rewriteIdentInFunc(d, renames)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if newName, ok := renames[s.Name.Name]; ok {
						s.Name.Name = newName
					}
				case *ast.ValueSpec:
					for i, n := range s.Names {
						if newName, ok := renames[n.Name]; ok {
							s.Names[i].Name = newName
						}
					}
				}
			}
		}
	}
	return decls
}

func (b *bundler) rewriteIdentInFunc(fn *ast.FuncDecl, renames map[string]string) {
	if fn.Body == nil {
		return
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if newName, ok2 := renames[ident.Name]; ok2 {
				ident.Name = newName
			}
		}
		return true
	})
}

func (b *bundler) rewriteFile(f *ast.File) {
	for _, decl := range f.Decls {
		b.rewriteDecl(decl)
	}
}

func (b *bundler) rewriteDecl(decl ast.Decl) {
	if decl == nil {
		return
	}
	switch d := decl.(type) {
	case *ast.FuncDecl:
		b.rewriteFieldList(d.Type.Params)
		b.rewriteFieldList(d.Type.Results)
		b.rewriteBlockStmt(d.Body)
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			if vs, ok := spec.(*ast.ValueSpec); ok {
				vs.Type = b.rewriteExpr(vs.Type)
				for i, v := range vs.Values {
					vs.Values[i] = b.rewriteExpr(v)
				}
			}
		}
	}
}

func (b *bundler) rewriteFieldList(fl *ast.FieldList) {
	if fl == nil {
		return
	}
	for _, field := range fl.List {
		field.Type = b.rewriteExpr(field.Type)
	}
}

func (b *bundler) rewriteBlockStmt(block *ast.BlockStmt) {
	if block == nil {
		return
	}
	for i, stmt := range block.List {
		block.List[i] = b.rewriteStmt(stmt)
	}
}

func (b *bundler) rewriteStmt(stmt ast.Stmt) ast.Stmt {
	if stmt == nil {
		return nil
	}
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for i, e := range s.Lhs {
			s.Lhs[i] = b.rewriteExpr(e)
		}
		for i, e := range s.Rhs {
			s.Rhs[i] = b.rewriteExpr(e)
		}
	case *ast.ExprStmt:
		s.X = b.rewriteExpr(s.X)
	case *ast.ReturnStmt:
		for i, e := range s.Results {
			s.Results[i] = b.rewriteExpr(e)
		}
	case *ast.IfStmt:
		s.Init = b.rewriteStmt(s.Init)
		s.Cond = b.rewriteExpr(s.Cond)
		b.rewriteBlockStmt(s.Body)
		if s.Else != nil {
			s.Else = b.rewriteStmt(s.Else)
		}
	case *ast.ForStmt:
		s.Init = b.rewriteStmt(s.Init)
		s.Cond = b.rewriteExpr(s.Cond)
		s.Post = b.rewriteStmt(s.Post)
		b.rewriteBlockStmt(s.Body)
	case *ast.RangeStmt:
		s.Key = b.rewriteExpr(s.Key)
		s.Value = b.rewriteExpr(s.Value)
		s.X = b.rewriteExpr(s.X)
		b.rewriteBlockStmt(s.Body)
	case *ast.BlockStmt:
		b.rewriteBlockStmt(s)
	case *ast.DeclStmt:
		b.rewriteDecl(s.Decl)
	case *ast.IncDecStmt:
		s.X = b.rewriteExpr(s.X)
	case *ast.SendStmt:
		s.Chan = b.rewriteExpr(s.Chan)
		s.Value = b.rewriteExpr(s.Value)
	case *ast.SwitchStmt:
		s.Init = b.rewriteStmt(s.Init)
		s.Tag = b.rewriteExpr(s.Tag)
		b.rewriteBlockStmt(s.Body)
	case *ast.TypeSwitchStmt:
		s.Init = b.rewriteStmt(s.Init)
		s.Assign = b.rewriteStmt(s.Assign)
		b.rewriteBlockStmt(s.Body)
	case *ast.CaseClause:
		for i, e := range s.List {
			s.List[i] = b.rewriteExpr(e)
		}
		for i, st := range s.Body {
			s.Body[i] = b.rewriteStmt(st)
		}
	case *ast.SelectStmt:
		b.rewriteBlockStmt(s.Body)
	case *ast.CommClause:
		s.Comm = b.rewriteStmt(s.Comm)
		for i, st := range s.Body {
			s.Body[i] = b.rewriteStmt(st)
		}
	case *ast.GoStmt:
		if e := b.rewriteExpr(s.Call); e != nil {
			s.Call = e.(*ast.CallExpr)
		}
	case *ast.DeferStmt:
		if e := b.rewriteExpr(s.Call); e != nil {
			s.Call = e.(*ast.CallExpr)
		}
	}
	return stmt
}

func (b *bundler) rewriteExpr(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			if renames, ok := b.pkgRename[ident.Name]; ok {
				if newName, ok := renames[e.Sel.Name]; ok {
					return &ast.Ident{Name: newName}
				}
				return &ast.Ident{Name: e.Sel.Name}
			}
		}
		e.X = b.rewriteExpr(e.X)
		return e
	case *ast.CallExpr:
		e.Fun = b.rewriteExpr(e.Fun)
		for i, arg := range e.Args {
			e.Args[i] = b.rewriteExpr(arg)
		}
		return e
	case *ast.BinaryExpr:
		e.X = b.rewriteExpr(e.X)
		e.Y = b.rewriteExpr(e.Y)
		return e
	case *ast.UnaryExpr:
		e.X = b.rewriteExpr(e.X)
		return e
	case *ast.IndexExpr:
		e.X = b.rewriteExpr(e.X)
		e.Index = b.rewriteExpr(e.Index)
		return e
	case *ast.SliceExpr:
		e.X = b.rewriteExpr(e.X)
		e.Low = b.rewriteExpr(e.Low)
		e.High = b.rewriteExpr(e.High)
		e.Max = b.rewriteExpr(e.Max)
		return e
	case *ast.CompositeLit:
		e.Type = b.rewriteExpr(e.Type)
		for i, elt := range e.Elts {
			e.Elts[i] = b.rewriteExpr(elt)
		}
		return e
	case *ast.KeyValueExpr:
		e.Key = b.rewriteExpr(e.Key)
		e.Value = b.rewriteExpr(e.Value)
		return e
	case *ast.TypeAssertExpr:
		e.X = b.rewriteExpr(e.X)
		e.Type = b.rewriteExpr(e.Type)
		return e
	case *ast.ParenExpr:
		e.X = b.rewriteExpr(e.X)
		return e
	case *ast.StarExpr:
		e.X = b.rewriteExpr(e.X)
		return e
	case *ast.FuncLit:
		b.rewriteFieldList(e.Type.Params)
		b.rewriteFieldList(e.Type.Results)
		b.rewriteBlockStmt(e.Body)
		return e
	}
	return expr
}

// rebuildImports は mainFile の import 宣言を allImports の内容で置き換える。
func rebuildImports(mainFile *ast.File, allImports map[string]bool) {
	newDecls := make([]ast.Decl, 0, len(mainFile.Decls))
	for _, decl := range mainFile.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if ok && gd.Tok == token.IMPORT {
			continue
		}
		newDecls = append(newDecls, decl)
	}
	if len(allImports) == 0 {
		mainFile.Decls = newDecls
		return
	}
	specs := make([]ast.Spec, 0, len(allImports))
	for imp := range allImports {
		specs = append(specs, &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: `"` + imp + `"`},
		})
	}
	importDecl := &ast.GenDecl{Tok: token.IMPORT, Lparen: 1, Specs: specs}
	mainFile.Decls = append([]ast.Decl{importDecl}, newDecls...)
	mainFile.Imports = nil
}

// findModule は dir から上位へ走査して go.mod を見つけ、ルートとモジュール名を返す。
func findModule(dir string) (root, name string, err error) {
	cur := dir
	for {
		data, readErr := os.ReadFile(filepath.Join(cur, "go.mod"))
		if readErr == nil {
			name = parseModuleName(string(data))
			if name == "" {
				return "", "", fmt.Errorf("go.mod に module 宣言がありません: %s", cur)
			}
			return cur, name, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", "", fmt.Errorf("go.mod が見つかりません（%s から上位を探索）", dir)
}

func parseModuleName(gomod string) string {
	for _, line := range strings.Split(gomod, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
