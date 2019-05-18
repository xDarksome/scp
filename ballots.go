package scp

import (
	"bytes"
	"time"

	"github.com/google/btree"
)

type ballot struct {
	slotIndex uint64
	counter   uint32
	value     Value

	prepare *ratification
	commit  *ratification
}

func (b *ballot) Less(than btree.Item) bool {
	return b.less(than.(*ballot))
}

func (b *ballot) less(ballot *ballot) bool {
	if b.slotIndex < ballot.slotIndex {
		return true
	}

	if b.slotIndex > ballot.slotIndex {
		return false
	}

	if b.counter < ballot.counter {
		return true
	}

	if b.counter > ballot.counter {
		return false
	}

	return bytes.Compare(b.value, ballot.value) == -1
}

func (b *ballot) compatible(ballot *ballot) bool {
	return bytes.Compare(b.value, ballot.value) == 0
}

type ballots struct {
	btree                    *btree.BTree
	current                  *ballot
	highestConfirmedPrepared *ballot
	highestAcceptedPrepared  *ballot
	counters                 ballotCounters
	timer                    *time.Timer
}

func newBallots(slotIndex uint64) ballots {
	return ballots{
		btree:    btree.New(100),
		current:  &ballot{slotIndex: slotIndex, counter: 1},
		counters: make(ballotCounters),
		timer:    time.NewTimer(time.Hour),
	}
}

func (b ballots) FindOrCreate(slotIndex uint64, counter uint32, value Value, slices quorumSlices) *ballot {
	bal := &ballot{slotIndex: slotIndex, counter: counter, value: value}
	if item := b.btree.Get(bal); item != nil {
		return item.(*ballot)
	}

	bal.prepare = newRatification()
	bal.commit = newRatification()
	b.btree.ReplaceOrInsert(bal)
	return bal
}
