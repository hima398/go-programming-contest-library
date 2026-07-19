package graph

import "github.com/liyue201/gostl/ds/priorityqueue"

type DijkstraEdge struct {
	to     int
	weight int
}

type DijkstraNode struct {
	to   int
	cost int
}

type DijkstraGraph struct {
	edge       [][]DijkstraEdge
	isDirected bool
	dist       []int
	queue      *priorityqueue.PriorityQueue[DijkstraNode]
}

func NewDijkstraGraph(nodes int) *DijkstraGraph {
	res := new(DijkstraGraph)
	res.edge = make([][]DijkstraEdge, nodes)
	res.dist = make([]int, nodes)
	res.queue = priorityqueue.New[DijkstraNode](func(a, b DijkstraNode) int {
		if a.cost == b.cost {
			return 0
		} else if a.cost < b.cost {
			return -1
		} else {
			return 1
		}
	})
	return res
}

func (g *DijkstraGraph) AddEdge(from, to, weight int) {
	g.edge[from] = append(g.edge[from], DijkstraEdge{to, weight})
}

func (g *DijkstraGraph) GetDistance() []int {
	return g.dist
}

func (g *DijkstraGraph) ComputeShortestPath() {
	const INF = 1 << 60
	for i := range g.dist {
		g.dist[i] = INF
	}
	g.push(0, 0)
	for !g.queue.Empty() {
		cur := g.queue.Pop()
		if g.dist[cur.to] < cur.cost {
			continue
		}
		for _, next := range g.edge[cur.to] {
			g.push(next.to, cur.cost+next.weight)
		}
	}
}

func (g *DijkstraGraph) push(to, cost int) {
	if g.dist[to] <= cost {
		return
	}
	g.dist[to] = cost
	g.queue.Push(DijkstraNode{to, cost})
}
