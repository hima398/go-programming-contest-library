package util

import "testing"

func TestCompress_GetIndex(t *testing.T) {
	type fields struct {
		x []int
	}
	type args struct {
		x int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{"", fields{[]int{6, 8, 12}}, args{5}, 0},
		{"", fields{[]int{6, 8, 12}}, args{6}, 0},
		{"", fields{[]int{6, 8, 12}}, args{7}, 1},
		{"", fields{[]int{6, 8, 12}}, args{8}, 1},
		{"", fields{[]int{6, 8, 12}}, args{9}, 2},
		{"", fields{[]int{6, 8, 12}}, args{11}, 2},
		{"", fields{[]int{6, 8, 12}}, args{12}, 2},
		{"", fields{[]int{6, 8, 12}}, args{13}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compress{
				x: tt.fields.x,
			}
			if got := c.GetIndex(tt.args.x); got != tt.want {
				t.Errorf("Compress.GetIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
