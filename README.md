# go-programming-contest-library

AtCoder などの競技プログラミング向け Go ライブラリ集です。
開発時は各パッケージを `import` して使い、提出前に `bundle` コマンドで単一ファイルにまとめます。

---

## bundle コマンドのインストール

```bash
go install github.com/hima398/go-programming-contest-library/cmd/bundle@latest
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
bundle -in contest/abc300/main.go -out contest/abc300/submit.go
# または
make bundle CONTEST=contest/abc300
```

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

bundle を実行します。

```bash
bundle -in abc300/main.go -out abc300/submit.go
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
