package spanningtree

type SpanningTreeEdge struct {
	s, t int
	w    int
}

type SpanningTree struct {
	n int
	e [][]SpanningTreeEdge
}

func NewSpanningTree(n int) *SpanningTree {
	t := new(SpanningTree)
	t.n = n
	t.e = make([][]SpanningTreeEdge, n)

	return t
}

func (t *SpanningTree) Add(u, v, w int) {
	t.e[u] = append(t.e[u], SpanningTreeEdge{u, v, w})
	t.e[v] = append(t.e[v], SpanningTreeEdge{v, u, w})
}

func (t *SpanningTree) Diameter() int {
	dist1 := t.computeDistance(0)
	maxDist := 0
	s := 0
	for i, v := range dist1 {
		if maxDist < v {
			s = i
			maxDist = v
		}
	}
	dist2 := t.computeDistance(s)
	res := 0
	for _, v := range dist2 {
		if res < v {
			res = v
		}
	}

	return res
}

// xを視点として書くノードへの距離を求める
func (t *SpanningTree) computeDistance(x int) []int {
	dist := make([]int, t.n)
	for i := range dist {
		dist[i] = -1
	}
	var q []int
	q = append(q, x)
	dist[x] = 0
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, next := range t.e[cur] {
			//親ノード(訪問済み)
			if dist[next.t] >= 0 {
				continue
			}
			q = append(q, next.t)
			dist[next.t] = dist[cur] + next.w
		}
	}
	return dist
}
