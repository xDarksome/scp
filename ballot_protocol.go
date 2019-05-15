package scp

import (
	"bytes"
	"fmt"
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

type ballotProtocol struct {
	slotIndex uint64

	current                  *ballot
	highestConfirmedPrepared *ballot
	highestAcceptedPrepared  *ballot
	candidate                Value

	quorumSlices
	ballots
	counters ballotCounters

	timer      *time.Timer
	input      chan *Message
	candidates chan Value

	network Network
	slots   chan Slot
	stop    chan struct{}
}

func newBallotProtocol(slot uint64, slices quorumSlices, network Network) *ballotProtocol {
	return &ballotProtocol{
		slotIndex:    slot,
		current:      &ballot{slotIndex: slot, counter: 1},
		quorumSlices: slices,
		ballots:      newBallots(),
		counters:     make(ballotCounters),
		timer:        time.NewTimer(time.Hour),
		input:        make(chan *Message, 1000000),
		candidates:   make(chan Value, 1000),
		network:      network,
		slots:        make(chan Slot),
		stop:         make(chan struct{}),
	}
}

func (p *ballotProtocol) Run() {
	for {
		select {
		case <-p.timer.C:
			p.updateLocalCounter(p.current.counter + 1)
		case c := <-p.candidates:
			p.newCandidate(c)
		case m := <-p.input:
			p.receiveMessage(m)
		case <-p.stop:
			return
		}
	}
}

func (p *ballotProtocol) Stop() {
	p.stop <- struct{}{}
}

func (p *ballotProtocol) newCandidate(value Value) {
	p.candidate = value

	if p.current.value != nil {
		return
	}

	p.current.value = p.candidate
	p.votePrepare()
}

func (p *ballotProtocol) receiveMessage(m *Message) {
	b := &ballot{slotIndex: m.SlotIndex, counter: m.Counter, value: m.Value}
	switch m.Type {
	case VotePrepare:
		p.prepareVoted(m.NodeID, b)
	case AcceptPrepare:
		p.prepareAccepted(m.NodeID, b)
	case VoteCommit:
		p.commitVoted(m.NodeID, b)
	case AcceptCommit:
		p.commitAccepted(m.NodeID, b)
	}
}

func (p *ballotProtocol) newTimer() {
	p.timer.Stop()
	p.timer = time.NewTimer(time.Duration(p.current.counter+1) * time.Second)
}

func (p *ballotProtocol) recomputeLocalValue() {
	if p.highestConfirmedPrepared != nil {
		p.current.value = p.highestConfirmedPrepared.value
		return
	}

	if p.candidate != nil {
		p.current.value = p.candidate
		return
	}

	if p.highestAcceptedPrepared != nil {
		p.current.value = p.highestAcceptedPrepared.value
		return
	}
}

func (p *ballotProtocol) updateLocalCounter(c uint32) {
	p.current.counter = c
	p.recomputeLocalValue()
	p.votePrepare()
}

func (p *ballotProtocol) updateCounter(nodeID string, c uint32) {
	current := p.counters[nodeID]
	if current >= c {
		return
	}

	p.counters[nodeID] = c
	lowest := p.counters.lowestNonBlocking(p.current.counter, p.quorumSlices)
	if lowest > p.current.counter {
		p.updateLocalCounter(lowest)
	}

	if !p.counters.hasLesser(p.current.counter) {
		p.newTimer()
	}
}

func (p *ballotProtocol) prepareVoted(nodeID string, ballot *ballot) {
	fmt.Println(p.network.ID(), "see", nodeID, "has voted prepare", ballot.value)
	if ballot.Less(p.current) {
		return
	}

	p.updateCounter(nodeID, ballot.counter)
	ballot = p.ballots.GetOrInsert(ballot, p.quorumSlices)
	ballot.prepare.votedBy(nodeID)

	if !ballot.prepare.accepted && ballot.prepare.votes.reachedQuorumThreshold(p.quorumSlices) {
		p.acceptPrepare(ballot)
	}
}

func (p *ballotProtocol) prepareAccepted(nodeID string, ballot *ballot) {
	fmt.Println(p.network.ID(), "see", nodeID, "has accepted prepare", ballot.value)
	p.updateCounter(nodeID, ballot.counter)
	ballot = p.ballots.GetOrInsert(ballot, p.quorumSlices)
	ballot.prepare.acceptedBy(nodeID)

	if !ballot.prepare.accepted && ballot.prepare.accepts.reachedBlockingThreshold(p.quorumSlices) {
		p.acceptPrepare(ballot)
	}

	if !ballot.prepare.confirmed && ballot.prepare.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.confirmPrepare(ballot)
	}
}

func (p *ballotProtocol) commitVoted(nodeID string, ballot *ballot) {
	fmt.Println(p.network.ID(), "see", nodeID, "has voted commit", ballot.value)
	p.updateCounter(nodeID, ballot.counter)
	ballot = p.ballots.GetOrInsert(ballot, p.quorumSlices)
	ballot.commit.votedBy(nodeID)

	if !ballot.commit.accepted && ballot.commit.votes.reachedQuorumThreshold(p.quorumSlices) {
		p.acceptCommit(ballot)
	}
}

func (p *ballotProtocol) commitAccepted(nodeID string, ballot *ballot) {
	fmt.Println(p.network.ID(), "see", nodeID, "has accepted commit", ballot.value)
	p.updateCounter(nodeID, ballot.counter)
	ballot = p.ballots.GetOrInsert(ballot, p.quorumSlices)
	ballot.commit.acceptedBy(nodeID)

	if !ballot.commit.accepted && ballot.commit.accepts.reachedBlockingThreshold(p.quorumSlices) {
		p.acceptCommit(ballot)
	}

	if !ballot.commit.confirmed && ballot.commit.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.confirmCommit(ballot)
	}
}

func (p *ballotProtocol) votePrepare() {
	ballot := p.ballots.GetOrInsert(p.current, p.quorumSlices)
	p.current = ballot
	fmt.Println(p.network.ID(), "votes for prepare", ballot.value)

	if ballot.slotIndex == p.slotIndex {
		p.network.Broadcast(&Message{
			Type:      VotePrepare,
			SlotIndex: ballot.slotIndex,
			Counter:   ballot.counter,
			Value:     ballot.value,
		})
	}
	ballot.prepare.selfVote()

	if !ballot.prepare.accepted && ballot.prepare.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.acceptPrepare(ballot)
	}
}

func (p *ballotProtocol) acceptPrepare(ballot *ballot) {
	fmt.Println(p.network.ID(), "accepts prepare", ballot.value)
	if ballot.slotIndex == p.slotIndex {
		p.network.Broadcast(&Message{
			Type:      AcceptPrepare,
			SlotIndex: ballot.slotIndex,
			Counter:   ballot.counter,
			Value:     ballot.value,
		})
	}
	ballot.prepare.selfAccept()

	if p.highestAcceptedPrepared == nil || p.highestAcceptedPrepared.less(ballot) {
		p.highestAcceptedPrepared = ballot
	}

	if !ballot.prepare.confirmed && ballot.prepare.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.confirmPrepare(ballot)
	}
}

func (p *ballotProtocol) confirmPrepare(ballot *ballot) {
	fmt.Println(p.network.ID(), "confirms prepare", ballot.value)
	if p.highestConfirmedPrepared == nil || p.highestConfirmedPrepared.less(ballot) {
		p.highestConfirmedPrepared = ballot
	}

	ballot.prepare.selfConfirm()
	if !ballot.less(p.current) || ballot.compatible(p.current) {
		p.voteCommit(ballot)
	}
}

func (p *ballotProtocol) voteCommit(ballot *ballot) {
	fmt.Println(p.network.ID(), "votes for commit", ballot.value)
	if ballot.slotIndex == p.slotIndex {
		p.network.Broadcast(&Message{
			Type:      VoteCommit,
			SlotIndex: ballot.slotIndex,
			Counter:   ballot.counter,
			Value:     ballot.value,
		})
	}
	ballot.commit.selfVote()

	if !ballot.commit.accepted && ballot.commit.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.acceptCommit(ballot)
	}
}

func (p *ballotProtocol) acceptCommit(ballot *ballot) {
	fmt.Println(p.network.ID(), "accepts commit", ballot.value)
	if ballot.slotIndex == p.slotIndex {
		p.network.Broadcast(&Message{
			Type:      AcceptCommit,
			SlotIndex: ballot.slotIndex,
			Counter:   ballot.counter,
			Value:     ballot.value,
		})
	}
	ballot.commit.selfAccept()

	if !ballot.commit.confirmed && ballot.commit.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.confirmCommit(ballot)
	}
}

func (p *ballotProtocol) confirmCommit(ballot *ballot) {
	fmt.Println(p.network.ID(), "confirms commit", ballot.value)
	ballot.commit.selfConfirm()
	p.slots <- Slot{
		Index: p.slotIndex,
		Value: ballot.value,
	}
	p.slotIndex++
}
