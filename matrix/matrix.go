package matrix

type Matrix struct {
	r, c int
	data [][]int
}

func NewMatrix(r, c int) *Matrix {
	res := new(Matrix)
	res.r = r
	res.c = c
	res.data = make([][]int, r)
	for i := range res.data {
		res.data[i] = make([]int, c)
	}

	return res
}

func (m *Matrix) Get(i, j int) int {
	return m.data[i][j]
}

func (m *Matrix) Set(i, j, x int) {
	m.data[i][j] = x
}

func (m1 *Matrix) ModMul(m2 *Matrix, p int) *Matrix {
	res := NewMatrix(m1.r, m2.c)

	for i := range res.r {
		for j := range res.c {
			for k := range m1.c {
				res.data[i][j] += m1.data[i][k] * m2.data[k][j]
				res.data[i][j] %= p
			}
		}
	}
	return res
}

func (m *Matrix) ModPow(x, p int) *Matrix {
	res := NewMatrix(m.r, m.c)
	var n int
	if m.r < m.c {
		n = m.r
	} else {
		n = m.c
	}
	for i := range n {
		res.data[i][i] = 1
	}
	for x > 0 {
		if x%2 == 1 {
			res = res.ModMul(m, p)
		}
		x >>= 1
		m = m.ModMul(m, p)
	}

	return res
}
