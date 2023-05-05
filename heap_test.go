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

		want = []string{"e", "a", "b", "c", "d"}
	)

	p.add("b", exptimeB)
	p.add("a", exptimeA)
	p.add("c", exptimeC)
	p.add("d", exptimeD)
	p.add("e", exptimeE)

	heap.Init(&p)
	for _, expect := range want {
		res := heap.Pop(&p).(item)
		if res.key != expect {
			t.Fatalf("TestPriorityQueue: want %v got %v", expect, res.key)
		}
	}
}
