package main

import (
	"fmt"

	"github.com/hima398/go-programming-contest-library/dsu"
	"github.com/hima398/go-programming-contest-library/fenwicktree"
)

func main() {
	// DSU の使用例
	uf := dsu.New(5)
	uf.Unite(0, 1)
	uf.Unite(2, 3)
	fmt.Println(uf.ExistSameUnion(0, 1)) // true
	fmt.Println(uf.ExistSameUnion(0, 2)) // false

	// FenwickTree の使用例
	ft := fenwicktree.New(5)
	ft.Update(0, 10)
	ft.Update(1, 20)
	ft.Update(2, 30)
	fmt.Println(ft.Sum(0, 3)) // 60
}
