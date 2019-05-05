package quorum

import (
	"bytes"

	"github.com/xdarksome/scp/pkg/network"
)

type Agreement struct {
	Subject *network.Value

	Vote     *Threshold
	Accepted bool

	Accept    *Threshold
	Confirmed bool
}

func AgreementOn(segment *network.Value) *Agreement {
	return &Agreement{
		Subject: segment,
	}
}

func (a *Agreement) WithThresholds(self string, s Slice) *Agreement {
	a.Vote = newThreshold(self, s)
	a.Accept = newThreshold(self, s)
	return a
}

func (a *Agreement) VotedBy(id string) {
	a.Vote.NodeAgreed(id)
}

func (a *Agreement) AcceptedBy(id string) {
	a.VotedBy(id)
	a.Accept.NodeAgreed(id)
}

type Threshold struct {
	self   bool
	selfID string

	total    uint32
	expected uint32
	current  uint32

	pending      map[string]struct{}
	pendingInner []*innerThreshold
}

func (t *Threshold) ReachedQuorumThreshold() bool {
	return t.self && t.current >= t.expected
}

func (t *Threshold) ReachedBlockingThreshold() bool {
	return t.current >= t.total-t.expected
}

func (t *Threshold) SelfAgreedWith(nodeID string) {
	if nodeID != t.selfID {
		t.NodeAgreed(nodeID)
	}
	t.NodeAgreed(t.selfID)
}

func (t *Threshold) NodeAgreed(id string) {
	if id == t.selfID {
		t.self = true
	}

	if _, ok := t.pending[id]; ok {
		t.current++
		delete(t.pending, id)
	}

	for i, inner := range t.pendingInner {
		inner.nodeAgreed(id)
		if !inner.reachedQuorumThreshold() {
			continue
		}

		t.current++
		t.pendingInner[i] = t.pendingInner[len(t.pendingInner)-1]
		t.pendingInner[len(t.pendingInner)-1] = nil
		t.pendingInner = t.pendingInner[:len(t.pendingInner)-1]
	}
}

type innerThreshold struct {
	total    uint32
	expected uint32
	current  uint32
	pending  map[string]struct{}
}

func (a *Agreement) Less(b *Agreement) bool {
	return bytes.Compare(a.Subject.Data, b.Subject.Data) == -1
}

func (i *innerThreshold) nodeAgreed(id string) {
	if _, ok := i.pending[id]; ok {
		i.current++
		delete(i.pending, id)
	}
}

func (i *innerThreshold) reachedQuorumThreshold() bool {
	return i.current >= i.expected
}

func (i *innerThreshold) reachedBlockingThreshold() bool {
	return i.current >= i.total-i.expected
}

func newThreshold(self string, s Slice) *Threshold {
	t := Threshold{}
	t.selfID = self
	t.expected = s.Threshold
	t.pending = map[string]struct{}{}
	for _, id := range s.Validators {
		t.pending[id] = struct{}{}
	}

	for _, inner := range s.InnerSlices {
		it := innerThreshold{
			pending: map[string]struct{}{},
		}
		for _, id := range inner.Validators {
			it.pending[id] = struct{}{}
		}
		t.pendingInner = append(t.pendingInner, &it)
	}

	t.total = uint32(len(t.pending) + len(t.pendingInner))

	return &t
}
