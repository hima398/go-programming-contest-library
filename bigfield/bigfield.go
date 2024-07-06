package bigfield

//TODO:ABC331Dから持ってきただけなので要改善

// 縦横最大1e9のグリッドを表す構造体
type field struct {
	n int
	p []string

	s [][]int
}

func NewField(n int, p []string) *field {
	res := new(field)
	res.n = n
	res.p = p

	res.s = make([][]int, res.n)
	for i := range res.s {
		res.s[i] = make([]int, res.n)
	}
	for i := range res.s {
		for j := range res.s[i] {
			if p[i][j] == 'B' {
				res.s[i][j]++
			}
		}
	}
	for i := range res.s {
		for j := 0; j < res.n-1; j++ {
			res.s[i][j+1] += res.s[i][j]
		}
	}
	for j := 0; j < res.n; j++ {
		for i := 0; i < res.n-1; i++ {
			res.s[i+1][j] += res.s[i][j]
		}
	}
	return res
}

// (0, 0)を左上、(i, j)を右下とする矩形内の黒色のマスの数を計算する
// 計算量はO(1)
func (f *field) cumulativeSum(i, j int) int {
	di, dj := i/f.n, j/f.n
	mi, mj := i%f.n, j%f.n
	res := di*dj*f.s[f.n-1][f.n-1] + di*f.s[f.n-1][mj] + dj*f.s[mi][f.n-1] + f.s[mi][mj]
	return res
}

// (a, b)を左上、(c, d)を右下とする矩形内の黒色のマスの数を計算する。
// 計算量はO(1)
func (f *field) CountBlackCells(a, b, c, d int) int {
	var res int
	if a == 0 && b == 0 {
		return f.cumulativeSum(c, d)
	} else if a == 0 {
		res = f.cumulativeSum(c, d) - f.cumulativeSum(c, b-1)
	} else if b == 0 {
		res = f.cumulativeSum(c, d) - f.cumulativeSum(a-1, d)
	} else {
		res = f.cumulativeSum(c, d) - f.cumulativeSum(a-1, d) - f.cumulativeSum(c, b-1) + f.cumulativeSum(a-1, b-1)
	}
	return res
}
