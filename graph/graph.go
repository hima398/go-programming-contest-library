package graph

import (
	"container/heap"
)

const INF = 1 << 60

type Graph struct {
	n, m int
	d    []int    //distance
	v    []bool   // visited
	e    [][]Edge //edges
}

func New(n, m int) *Graph {
	return NewGraph(n, m)
}

func NewGraph(n, m int) *Graph {
	g := new(Graph)
	g.n = n
	g.m = m
	g.d = make([]int, n)
	for i := 0; i < n; i++ {

	}
	g.v = make([]bool, n)
	g.e = make([][]Edge, n)
	return g
}

func (g *Graph) Add(a, b, w int) {
	g.e[a] = append(g.e[a], Edge{a, b, w})
	g.e[b] = append(g.e[b], Edge{a, b, w})
}

func (g *Graph) Dfs(cur int) {
	g.v[cur] = true
	for _, next := range g.e[cur] {
		if g.v[next.t] {
			continue
		}
		g.Dfs(next.t)
	}
}

func (g *Graph) Bfs(s int) {
	var q []int
	q = append(q, s)
	g.v[s] = true
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, next := range g.e[cur] {
			if g.v[next.t] {
				continue
			}
			q = append(q, next.t)
			g.v[next.t] = true
		}
	}
}

func (g *Graph) Dijkstra(s, t int) int {
	q := &PriorityQueue{}
	heap.Init(q)

	push := func(i, t, c int) {
		if g.d[t] > c {
			g.d[t] = c
			heap.Push(q, Edge{i, t, c})
		}
	}
	push(0, 0, 0)
	g.d[s] = 0

	for q.Len() > 0 {
		cur := heap.Pop(q).(Edge)
		if g.d[cur.t] < cur.w {
			continue
		}
		for _, next := range g.e[cur.t] {
			push(next.s, next.t, cur.w+next.w)
		}
	}
	return g.d[t]
}

func (g *Graph) get

func Dijkstra(graph Interface) {
	q := &PriorityQueue{}
	heap.Init(q)

	push := func(i, t, c int) {
		if g.d[t] > c {
			g.d[t] = c
			heap.Push(q, Edge{i, t, c})
		}
	}
	push(0, 0, 0)
	g.d[s] = 0

	for q.Len() > 0 {
		cur := heap.Pop(q).(Edge)
		if g.d[cur.t] < cur.w {
			continue
		}
		for _, next := range g.e[cur.t] {
			push(next.s, next.t, cur.w+next.w)
		}
	}
	return g.d[t]

}

type Edge struct {
	s, t int
	w    int
}

type PriorityQueue []Edge

func (pq PriorityQueue) Len() int {
	return len(pq)
}
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].w < pq[j].w
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(item interface{}) {
	*pq = append(*pq, item.(Edge))
}

func (pq *PriorityQueue) Pop() interface{} {
	es := *pq // EdgeのSlice
	n := len(es)
	item := es[n-1]
	*pq = es[0 : n-1]
	return item
}
