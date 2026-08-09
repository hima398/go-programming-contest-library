package matrix

import "testing"

func newMatrixFromData(data [][]int) *Matrix {
	return &Matrix{r: len(data), c: len(data[0]), data: data}
}

func TestMatrix_ModMul(t *testing.T) {
	tests := []struct {
		name string
		m1   [][]int
		m2   [][]int
		p    int
		want [][]int
	}{
		{
			name: "square 2x2",
			m1:   [][]int{{1, 2}, {3, 4}},
			m2:   [][]int{{5, 6}, {7, 8}},
			p:    1000,
			want: [][]int{{19, 22}, {43, 50}},
		},
		{
			name: "rectangular 2x3 * 3x2",
			m1:   [][]int{{1, 2, 3}, {4, 5, 6}},
			m2:   [][]int{{7, 8}, {9, 10}, {11, 12}},
			p:    1000,
			want: [][]int{{58, 64}, {139, 154}},
		},
		{
			name: "mod wraparound 1x1",
			m1:   [][]int{{3}},
			m2:   [][]int{{4}},
			p:    5,
			want: [][]int{{2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m1 := newMatrixFromData(tt.m1)
			m2 := newMatrixFromData(tt.m2)
			got := m1.ModMul(m2, tt.p)
			if got.r != len(tt.want) || got.c != len(tt.want[0]) {
				t.Fatalf("ModMul() size = %dx%d, want %dx%d", got.r, got.c, len(tt.want), len(tt.want[0]))
			}
			for i := range got.data {
				for j := range got.data[i] {
					if got.data[i][j] != tt.want[i][j] {
						t.Errorf("ModMul() data[%d][%d] = %v, want %v", i, j, got.data[i][j], tt.want[i][j])
					}
				}
			}
		})
	}
}
