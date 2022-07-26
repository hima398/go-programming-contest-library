package fenwicktree

import "testing"

func TestFenwickTree_Query(t *testing.T) {
	type fields struct {
		n     int
		nodes []int
	}
	type args struct {
		i int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{"Empty", fields{1, []int{0}}, args{0}, 0},
		{"Size 1", fields{2, []int{0, 1}}, args{1}, 1},
		{"Size 4, [0, 1]", fields{4, []int{0, 1, 2, 1, 4}}, args{1}, 1},
		{"Size n, [0, 2]", fields{4, []int{0, 1, 2, 1, 4}}, args{2}, 2},
		{"Size n, [0, 3]", fields{4, []int{0, 1, 2, 1, 4}}, args{3}, 3},
		{"Size n, [0, 4]", fields{4, []int{0, 1, 2, 1, 4}}, args{4}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fen := &FenwickTree{
				n:     tt.fields.n,
				nodes: tt.fields.nodes,
			}
			if got := fen.Query(tt.args.i); got != tt.want {
				t.Errorf("FenwickTree.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}
