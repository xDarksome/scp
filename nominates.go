package scp

import (
	"bytes"

	"github.com/google/btree"
)

type nominate struct {
	value Value
	*ratification
}

func (n *nominate) Less(than btree.Item) bool {
	return bytes.Compare(n.value, than.(*nominate).value) == -1
}

type nominates struct {
	btree *btree.BTree
}

func newNominates() nominates {
	return nominates{btree.New(100)}
}

func (n nominates) FindOrCreate(value Value, slices quorumSlices) *nominate {
	nom := n.Find(value)
	if nom == nil {
		nom = n.Create(value, slices)
	}

	return nom
}

func (n nominates) Find(value Value) *nominate {
	if item := n.btree.Get(&nominate{value: value}); item != nil {
		return item.(*nominate)
	}

	return nil
}

func (n nominates) Create(value Value, slices quorumSlices) *nominate {
	nominate := &nominate{
		value:        value,
		ratification: newRatification(),
	}
	n.btree.ReplaceOrInsert(nominate)
	return nominate
}
