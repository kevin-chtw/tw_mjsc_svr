package ai

import (
	"container/heap"
	"math"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

type experience struct {
	state    []float32
	target   []float32
	priority float64
}
type priorityQueue []*experience

func (pq priorityQueue) Len() int { return len(pq) }

// 【修复】改为最小堆：优先级小的在堆顶，容量满时优先驱逐低优先级样本
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
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

// 【优化】批量采样：采样指定数量的高优先级样本，并保留部分低优先级样本在池中
func (p *PER) Sample() (states, targets [][]float32) {
	// 采样策略：取出所有样本，按优先级从高到低排序，取前N个用于训练，剩余的放回
	// 为简化实现，这里先取出所有样本
	allExps := make([]*experience, 0, len(p.pq))
	for len(p.pq) > 0 {
		exp := heap.Pop(&p.pq).(*experience)
		allExps = append(allExps, exp)
	}

	// 【优化】取出所有样本用于训练，训练后自动清空PER
	// 由于是最小堆，Pop出来的顺序是从小到大（低优先级在前）
	// 反转数组，使高优先级在前
	for i, j := 0, len(allExps)-1; i < j; i, j = i+1, j-1 {
		allExps[i], allExps[j] = allExps[j], allExps[i]
	}

	// 使用所有样本进行训练
	for _, exp := range allExps {
		states = append(states, exp.state)
		targets = append(targets, exp.target)
	}

	return
}

// Len 返回当前经验池样本数
func (p *PER) Len() int { return len(p.pq) }

func actionIndex(op int, tile int) int {
	switch op {
	case mahjong.OperateDiscard:
		return tile
	case mahjong.OperatePon:
		return 34 + tile
	case mahjong.OperateKon:
		return 68 + tile
	case mahjong.OperateHu:
		return 102 + tile
	case mahjong.OperatePass:
		return 136
	default:
		return 136
	}
}
