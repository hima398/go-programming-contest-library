package util

import "sort"

type Compress struct {
	//重複除去済みの圧縮元
	x []int
}

func New() *Compress {
	return NewCompress()
}

func NewCompress() *Compress {
	return new(Compress)
}

func (c *Compress) Init(x []int) {
	m := make(map[int]struct{})
	for _, v := range x {
		m[v] = struct{}{}
	}
	for k := range m {
		c.x = append(c.x, k)
	}
	sort.Ints(c.x)
}

func (c *Compress) GetIndex(x int) int {
	return sort.Search(len(c.x), func(i int) bool {
		return c.x[i] >= x
	})
}

func (c *Compress) Size() int {
	return len(c.x)
}
