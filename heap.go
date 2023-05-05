package redgla

import (
	"container/heap"
	"time"
)

// priorityQueue sorts the nodes in order of fastest response time.
type priorityQueue []item

type item struct {
	key   string
	spent time.Duration
}

func (p *priorityQueue) add(key string, spent time.Duration) {
	heap.Push(p, item{key, spent})
}

func (p priorityQueue) keys() (res []string) {
	res = make([]string, 0, len(p))
	for _, item := range p {
		res = append(res, item.key)
	}
	return
}

// heap.Interface boilerplate
func (p priorityQueue) Len() int            { return len(p) }
func (p priorityQueue) Less(i, j int) bool  { return p[i].spent < p[j].spent }
func (p priorityQueue) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *priorityQueue) Push(x interface{}) { *p = append(*p, x.(item)) }
func (p *priorityQueue) Pop() interface{} {
	old := *p
	n := len(old)
	x := old[n-1]
	old[n-1] = item{}
	*p = old[0 : n-1]
	return x
}
