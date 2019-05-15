package quorum

import (
	"bytes"

	"github.com/google/btree"
)

type NominatesBTree struct {
	btree.BTree
	slices *Slices
}

func NewNominatesBTree(s *Slices) *NominatesBTree {
	return &NominatesBTree{
		slices: s,
		BTree:  *btree.New(100),
	}
}

func (n *NominatesBTree) GetOrInsert(value []byte) *Nominate {
	if item := n.BTree.Get(&Nominate{Value: value}); item != nil {
		return item.(*Nominate)
	}

	nominate := newNominate(value, n.slices)
	n.ReplaceOrInsert(nominate)
	return nominate
}

func (n *NominatesBTree) Get(value []byte) *Nominate {
	item := n.BTree.Get(&Nominate{Value: value})
	if item == nil {
		return nil
	}

	return item.(*Nominate)
}

func (n *NominatesBTree) Insert(value []byte) *Nominate {
	nominate := newNominate(value, n.slices)
	n.ReplaceOrInsert(nominate)
	return nominate
}

type Nominate struct {
	Value []byte
	Voted bool

	Votes    thresholds
	Accepted bool

	Accepts   thresholds
	Confirmed bool
}

func newNominate(value []byte, slices *Slices) *Nominate {
	return &Nominate{
		Value:   value,
		Votes:   newThresholds(slices),
		Accepts: newThresholds(slices),
	}
}

func (n *Nominate) Less(than btree.Item) bool {
	return bytes.Compare(n.Value, than.(*Nominate).Value) == -1
}

func (n *Nominate) Vote() {
	n.Votes.self()
	n.Voted = true
}

func (n *Nominate) Accept() {
	n.Accepts.self()
	n.Accepted = true
}

func (n *Nominate) Confirm() {
	n.Confirmed = true
}

func (n *Nominate) VotedBy(id string) {
	n.Votes.mark(id)
}

func (n *Nominate) AcceptedBy(id string) {
	n.VotedBy(id)
	n.Accepts.mark(id)
}
