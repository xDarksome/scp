package quorum

import (
	"bytes"

	"github.com/google/btree"
)

type BallotsBTree struct {
	btree.BTree
	slices *Slices
}

func NewBallotsBTree() *BallotsBTree {
	return &BallotsBTree{
		BTree: *btree.New(100),
	}
}

func (bt *BallotsBTree) GetOrInsert(counter uint32, value []byte) *Ballot {
	b := &Ballot{Counter: counter, Value: value}

	if item := bt.Get(b); item != nil {
		return item.(*Ballot)
	}

	b.SetThresholds(bt.slices)
	bt.ReplaceOrInsert(b)
	return b
}

type Ballot struct {
	Counter uint32
	Value   []byte

	PrepareVoted bool

	PrepareVotes    thresholds
	PrepareAccepted bool

	PrepareAccepts   thresholds
	PrepareConfirmed bool

	CommitVoted bool

	CommitVotes    thresholds
	CommitAccepted bool

	CommitAccepts   thresholds
	CommitConfirmed bool
}

func (b *Ballot) Less(than btree.Item) bool {
	t := than.(*Ballot)

	if b.Counter < t.Counter {
		return true
	}

	if b.Counter > t.Counter {
		return false
	}

	return bytes.Compare(b.Value, t.Value) == -1
}

func (b *Ballot) SetThresholds(slices *Slices) *Ballot {
	b.PrepareVotes = newThresholds(slices)
	b.PrepareAccepts = newThresholds(slices)
	b.CommitVotes = newThresholds(slices)
	b.CommitAccepts = newThresholds(slices)
	return b
}

func (b *Ballot) VotePrepare() {
	b.PrepareVotes.self()
	b.PrepareVoted = true
}

func (b *Ballot) PrepareVotedBy(nodeID string) {
	b.PrepareVotes.mark(nodeID)
}

func (b *Ballot) AcceptPrepare() {
	b.PrepareAccepts.self()
	b.PrepareAccepted = true
}

func (b *Ballot) PrepareAcceptedBy(nodeID string) {
	b.PrepareAccepts.mark(nodeID)
}

func (b *Ballot) ConfirmPrepare() {
	b.PrepareConfirmed = true
}

func (b *Ballot) CommitVotedBy(nodeID string) {
	b.CommitVotes.mark(nodeID)
}

func (b *Ballot) CommitAcceptedBy(nodeID string) {
	b.CommitAccepts.mark(nodeID)
}
