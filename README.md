# go-programming-contest-library

AtCoder などの競技プログラミング向け Go ライブラリ集です。
開発時は各パッケージを `import` して使い、提出前に `ac-bundle` コマンドで単一ファイルにまとめます。

---

## ac-bundle コマンドのインストール

```bash
go install github.com/hima398/go-programming-contest-library/cmd/ac-bundle@latest
```

---

## 使い方

### パターン1: このリポジトリ内で開発する場合

`contest/<コンテスト名>/main.go` を作成し、ライブラリを import します。

```go
// contest/abc300/main.go
package main

import (
    "fmt"
    "github.com/hima398/go-programming-contest-library/dsu"
)

func main() {
    uf := dsu.New(5)
    uf.Unite(0, 1)
    fmt.Println(uf.ExistSameUnion(0, 1)) // true
}
```

提出用ファイルを生成します。

```bash
ac-bundle -in contest/abc300/main.go -out contest/abc300/submit.go
# または
make bundle CONTEST=contest/abc300
```

### AHC などでロジックをファイル分割したい場合

`ac-bundle` は `-in` に指定した1ファイルしか読みません。`main.go` と同じディレクトリに
`domain.go` のような兄弟ファイル（同じ `package main`）を置いて分割しても、それらは
import 文を経由しないため結合対象になりません。

ロジックを分割したい場合は、`main.go` のあるディレクトリの**サブフォルダ**にパッケージとして
切り出してください。`-in` 配下のサブフォルダは、自モジュールのパッケージとして inline 展開されます。

```
develop/
├── main.go          // package main、solver を import
└── solver/
    ├── domain.go    // package solver
    ├── io.go
    └── util.go
```

```go
// develop/main.go（go.mod の module が example.com/proj の場合）
import (
    "example.com/proj/develop/solver"
)
```

- ディレクトリ直下の全 `.go`（`_test.go` を除く）が1つのパッケージとして結合されます。
- パッケージ名はディレクトリ名と一致させてください。
- `main.go` から参照する識別子は export（大文字始まり）にしてください。
- 名前の候補: `solver` `sol` `core` `logic` `sim`。標準ライブラリ名（`io` `sort` など）や
  既存ライブラリパッケージ名（`util` `slice` など）は避けてください。

### ac-bundle が結合できる範囲

`ac-bundle` が inline 展開するのは、`main.go` が**直接** import している次のパッケージのみです。

- import path が `github.com/hima398/go-programming-contest-library/` から始まるパッケージ
- `-in` の `main.go` があるディレクトリ配下のサブフォルダ（`<module>/<main.goのディレクトリ>/...`）

| 対象 | 結合される？ |
|---|---|
| `dsu` などライブラリ直下のパッケージ | される |
| `-in` 配下のサブフォルダ（例: `develop/solver`） | される |
| 標準ライブラリ・外部モジュール（`fmt`, `golang.org/x/...` など） | されない（通常の import 文のまま残る） |
| `main.go` と同階層の兄弟ファイル（同一 `package main` への分割） | されない（import を経由しないため） |
| 展開したパッケージがさらに import するパッケージ | されない（展開は1段のみ） |

#### 制約

- `main.go` で import にエイリアスを付けない（`import d ".../dsu"` は不可）。
- 展開するパッケージの型を slice / map / chan 型の中や struct フィールド、レシーバに
  書くと置換されず、コンパイルエラーになります（関数の引数・戻り値、`var`、式の中は可）。
- `main.go` の標準ライブラリ import にエイリアスを付けない（展開時にエイリアスが消えます）。
- `main.go` のトップレベル名が展開パッケージの名前と衝突しないようにする。

### パターン2: 別の Go プロジェクトから使う場合

別プロジェクトの `go.mod` にこのライブラリを追加します。

```bash
# 別プロジェクトのルートで
go get github.com/hima398/go-programming-contest-library
```

ローカルチェックアウトを使う場合は `replace` ディレクティブを追加します。

```
// go.mod
require github.com/hima398/go-programming-contest-library v0.0.0

replace github.com/hima398/go-programming-contest-library => /path/to/local/checkout
```

ac-bundle を実行します。

```bash
ac-bundle -in abc300/main.go -out abc300/submit.go
```

---

## 利用可能なパッケージ

| パッケージ | 概要 |
|---|---|
| `dsu` | Union-Find（素集合データ構造） |
| `fenwicktree` | Fenwick Tree（Binary Indexed Tree） |
| `segmenttree` | セグメント木 |
| `lazysegmenttree` | 遅延セグメント木 |
| `prime` | エラトステネスの篩 |
| `util` | 座標圧縮 |
| `string` | ハミング距離 |
| `bigfield` | 大きなグリッド上の2次元累積和 |
| `avltree` | AVL木（`Put`/`Remove`/`IndexOf` は未実装） |
| `matrix` | 行列演算（mod付き乗算・累乗） |
| `numeric` | 数値計算ユーティリティ（Gcd/Lcm/Sqrt など） |
| `slice` | スライス操作ユーティリティ |
| `spanningtree` | 木の直径計算 |
