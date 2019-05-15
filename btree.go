package scp

import "github.com/google/btree"

type (
	nominates struct {
		*btree.BTree
	}
	ballots struct {
		*btree.BTree
	}
)

func newNominates() nominates {
	return nominates{btree.New(100)}
}

func newBallots() ballots {
	return ballots{btree.New(100)}
}

func (b ballots) GetOrInsert(bal *ballot, slices quorumSlices) *ballot {
	if item := b.Get(bal); item != nil {
		return item.(*ballot)
	}

	bal.prepare = newRatification()
	bal.commit = newRatification()
	b.ReplaceOrInsert(bal)
	return bal
}

func (n nominates) GetOrInsert(nominate *nominate, slices quorumSlices) *nominate {
	nom := n.Get(nominate)
	if nom == nil {
		nom = n.Insert(nominate, slices)
	}

	return nom
}

func (n nominates) Get(nom *nominate) *nominate {
	if item := n.BTree.Get(nom); item != nil {
		return item.(*nominate)
	}

	return nil
}

func (n nominates) Insert(nominate *nominate, slices quorumSlices) *nominate {
	nominate.ratification = newRatification()
	n.ReplaceOrInsert(nominate)
	return nominate
}
