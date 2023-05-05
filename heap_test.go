package redgla

import (
	"container/heap"
	"testing"
	"time"
)

func TestPriorityQueue(t *testing.T) {
	var p priorityQueue

	var (
		exptimeA = time.Duration(2 * time.Second)
		exptimeB = time.Duration(3 * time.Second)
		exptimeC = time.Duration(5 * time.Second)
		exptimeD = time.Duration(15 * time.Second)
		exptimeE = time.Duration(1 * time.Second)

		wants = []string{"e", "a", "b", "c", "d"}
	)

	p.add("b", exptimeB)
	p.add("a", exptimeA)
	p.add("c", exptimeC)
	p.add("d", exptimeD)
	p.add("e", exptimeE)

	heap.Init(&p)
	for _, want := range wants {
		res := heap.Pop(&p).(item)
		if res.key != want {
			t.Fatalf("TestPriorityQueue: want %v got %v", want, res.key)
		}
	}
}
