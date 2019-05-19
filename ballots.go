package scp

import (
	"bytes"

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
	btree *btree.BTree
}

func newBallots() ballots {
	return ballots{
		btree: btree.New(100),
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
