package segmenttree

type SegmentTree struct {
	size  int
	nodes []int
	f     func(x1, x2 int) int
	inf   int
}

func NewSegmentTree(n, inf int, f func(x1, x2 int) int) (st *SegmentTree) {
	st = new(SegmentTree)
	st.size = 1
	for st.size < n {
		st.size *= 2
	}
	st.nodes = make([]int, 2*st.size)
	for i := range st.nodes {
		st.nodes[i] = inf
	}
	st.inf = inf
	st.f = f
	return st
}

func (seg *SegmentTree) queryRecursively(a, b, k, l, r int) int {
	// [a, b)と[l, r)が交差しない
	if a >= r || b <= l {
		return seg.inf
	}

	// [a, b)が[l, r)を完全に含んでいる
	if a <= l && b >= r {
		return seg.nodes[k]
	}

	vl := seg.queryRecursively(a, b, 2*k, l, (l+r)/2)
	vr := seg.queryRecursively(a, b, 2*k+1, (l+r)/2, r)
	return seg.f(vl, vr) //Max(vl, vr)
}

func (seg *SegmentTree) Query(l, r int) int {
	return seg.queryRecursively(l, r, 1, 0, seg.size)
}

func (seg *SegmentTree) Update(k, x int) {
	k += seg.size
	seg.nodes[k] = x
	for k > 1 {
		k /= 2
		seg.nodes[k] = seg.f(seg.nodes[k*2], seg.nodes[k*2+1])
	}
}
