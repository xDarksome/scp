package scp

import "sync/atomic"

type slotIndex struct {
	atomic atomic.Value
}

type slots struct {
	currentIndex uint64
	output       chan Slot
}

func newSlots(currentIndex uint64) slots {
	return slots{
		currentIndex: currentIndex,
		output:       make(chan Slot, 1000),
	}
}
