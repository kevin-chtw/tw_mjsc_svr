package ai

import (
	"container/heap"
	"math"
)

type experience struct {
	state    []float32
	target   []float32
	priority float64
}
type priorityQueue []*experience

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority > pq[j].priority }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*experience))
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type PER struct {
	cap   int
	pq    priorityQueue
	alpha float64
}

func NewPER(cap int) *PER {
	return &PER{cap: cap, alpha: 0.6}
}

func (p *PER) Add(state, target []float32, tdErr float64) {
	prio := math.Pow(math.Abs(tdErr)+1e-6, p.alpha)
	exp := &experience{state: state, target: target, priority: prio}
	if len(p.pq) >= p.cap {
		heap.Pop(&p.pq)
	}
	heap.Push(&p.pq, exp)
}

func (p *PER) Sample(batch int) (states, targets [][]float32) {
	n := len(p.pq)
	if n < batch {
		batch = n
	}
	for i := 0; i < batch; i++ {
		exp := heap.Pop(&p.pq).(*experience)
		states = append(states, exp.state)
		targets = append(targets, exp.target)
		// 重新 push 保持堆大小
		heap.Push(&p.pq, exp)
	}
	return
}
