package scp

import (
	"time"
)

type SlotsLoader interface {
	LoadSlots(from uint64, to Slot) []Slot
}

type Ledger interface {
	PersistSlot(Slot)
}

type ballotProtocol struct {
	slotIndex uint64
	id        string

	loader SlotsLoader
	ledger Ledger

	quorumSlices
	ballots

	currentBallot            *ballot
	highestConfirmedPrepared *ballot
	highestAcceptedPrepared  *ballot
	counters                 ballotCounters
	timer                    *time.Timer

	inputMessages  chan *Message
	outputMessages chan *Message

	nominationProtocol chan protocolMessage
	lastCandidate      Value
	candidates         chan Value

	catchUpDone  chan struct{}
	persistLater []Slot
}

func (b *ballotProtocol) init(slotIndex uint64) {
	b.slotIndex = slotIndex
	b.ballots = newBallots()
	b.currentBallot = &ballot{slotIndex: slotIndex, counter: 1}
	b.counters = make(ballotCounters)
	b.timer = time.NewTimer(time.Hour)
}

func (b *ballotProtocol) reinit(index uint64) {
	b.init(index)

	b.candidates = make(chan Value, 1000)
	b.highestAcceptedPrepared = nil
	b.highestConfirmedPrepared = nil
	b.lastCandidate = nil

	b.nominationProtocol <- protocolMessage{
		slotIndex:  index,
		candidates: b.candidates,
	}
}

func (b *ballotProtocol) run() {
	for {
		select {
		case c := <-b.candidates:
			b.newCandidate(c)
		case m := <-b.inputMessages:
			b.receive(m)
		case <-b.timer.C:
			b.updateCurrentBallotCounter(b.currentBallot.counter + 1)
		case <-b.catchUpDone:
			b.finishCatchUp()
		}
	}
}

func (b *ballotProtocol) newCandidate(value Value) {
	first := b.lastCandidate == nil
	b.lastCandidate = value

	if first {
		b.currentBallot.value = value
		b.votePrepare()
	}
}

func (b *ballotProtocol) receive(m *Message) {
	if m.SlotIndex < b.slotIndex {
		return
	}

	switch m.Type {
	case VotePrepare:
		b.prepareVoted(m)
	case AcceptPrepare:
		b.prepareAccepted(m)
	case VoteCommit:
		b.commitVoted(m)
	case AcceptCommit:
		b.commitAccepted(m)
	}
}

func (b *ballotProtocol) updateCurrentBallotCounter(c uint32) {
	b.currentBallot.counter = c
	b.recomputeCurrentBallotValue()
	b.votePrepare()
}

func (b *ballotProtocol) broadcast(m *Message) {
	if m.SlotIndex != b.slotIndex {
		return
	}

	m.NodeID = b.id
	b.outputMessages <- m
}

func (b *ballotProtocol) prepareVoted(m *Message) {
	b.updateBallotCounter(m.NodeID, m.Counter)
	ballot := b.ballots.findOrCreate(m.SlotIndex, m.Counter, m.Value, b.quorumSlices)
	ballot.prepare.votedBy(m.NodeID)

	if !ballot.prepare.accepted && ballot.prepare.votes.reachedQuorumThreshold(b.quorumSlices) {
		b.acceptPrepare(ballot)
	}
}

func (b *ballotProtocol) prepareAccepted(m *Message) {
	b.updateBallotCounter(m.NodeID, m.Counter)
	ballot := b.ballots.findOrCreate(m.SlotIndex, m.Counter, m.Value, b.quorumSlices)
	ballot.prepare.acceptedBy(m.NodeID)

	if !ballot.prepare.accepted && ballot.prepare.accepts.reachedBlockingThreshold(b.quorumSlices) {
		b.acceptPrepare(ballot)
	}

	if !ballot.prepare.confirmed && ballot.prepare.accepts.reachedQuorumThreshold(b.quorumSlices) {
		b.confirmPrepare(ballot)
	}
}

func (b *ballotProtocol) commitVoted(m *Message) {
	b.updateBallotCounter(m.NodeID, m.Counter)
	ballot := b.ballots.findOrCreate(m.SlotIndex, m.Counter, m.Value, b.quorumSlices)
	ballot.commit.votedBy(m.NodeID)

	if !ballot.commit.accepted && ballot.commit.votes.reachedQuorumThreshold(b.quorumSlices) {
		b.acceptCommit(ballot)
	}
}

func (b *ballotProtocol) commitAccepted(m *Message) {
	b.updateBallotCounter(m.NodeID, m.Counter)
	ballot := b.ballots.findOrCreate(m.SlotIndex, m.Counter, m.Value, b.quorumSlices)
	ballot.commit.acceptedBy(m.NodeID)

	if !ballot.commit.accepted && ballot.commit.accepts.reachedBlockingThreshold(b.quorumSlices) {
		b.acceptCommit(ballot)
	}

	if !ballot.commit.confirmed && ballot.commit.accepts.reachedQuorumThreshold(b.quorumSlices) {
		b.confirmCommit(ballot)
	}
}

func (b *ballotProtocol) votePrepare() {
	ballot := b.ballots.findOrCreate(
		b.currentBallot.slotIndex,
		b.currentBallot.counter,
		b.currentBallot.value,
		b.quorumSlices)
	b.currentBallot = ballot

	b.broadcast(&Message{
		Type:      VotePrepare,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.prepare.selfVote()

	if !ballot.prepare.accepted && ballot.prepare.accepts.reachedQuorumThreshold(b.quorumSlices) {
		b.acceptPrepare(ballot)
	}
}

func (b *ballotProtocol) acceptPrepare(ballot *ballot) {
	b.broadcast(&Message{
		Type:      AcceptPrepare,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.prepare.selfAccept()

	if b.highestAcceptedPrepared == nil || b.highestAcceptedPrepared.less(ballot) {
		b.highestAcceptedPrepared = ballot
	}

	if !ballot.prepare.confirmed && ballot.prepare.accepts.reachedQuorumThreshold(b.quorumSlices) {
		b.confirmPrepare(ballot)
	}
}

func (b *ballotProtocol) confirmPrepare(ballot *ballot) {
	if b.highestConfirmedPrepared == nil || b.highestConfirmedPrepared.less(ballot) {
		b.highestConfirmedPrepared = ballot
	}

	ballot.prepare.selfConfirm()
	if !ballot.less(b.currentBallot) || ballot.compatible(b.currentBallot) {
		b.voteCommit(ballot)
	}
}

func (b *ballotProtocol) voteCommit(ballot *ballot) {
	b.broadcast(&Message{
		Type:      VoteCommit,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.commit.selfVote()

	if !ballot.commit.accepted && ballot.commit.accepts.reachedQuorumThreshold(b.quorumSlices) {
		b.acceptCommit(ballot)
	}
}

func (b *ballotProtocol) acceptCommit(ballot *ballot) {
	b.broadcast(&Message{
		Type:      AcceptCommit,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.commit.selfAccept()

	if !ballot.commit.confirmed && ballot.commit.accepts.reachedQuorumThreshold(b.quorumSlices) {
		b.confirmCommit(ballot)
	}
}

func (b *ballotProtocol) confirmCommit(ballot *ballot) {
	ballot.commit.selfConfirm()
	b.externalize(Slot{
		Index: ballot.slotIndex,
		Value: ballot.value,
	})
}

func (b *ballotProtocol) externalize(s Slot) {
	if s.Index != b.slotIndex {
		b.catchUp(s)
		b.reinit(s.Index)
		return
	}

	b.ledger.PersistSlot(s)
	b.reinit(s.Index + 1)
}

func (b *ballotProtocol) catchUp(slot Slot) {
	inProgress := b.persistLater != nil
	b.persistLater = append(b.persistLater, slot)

	if inProgress {
		return
	}

	go catchuper{
		loader: b.loader,
		ledger: b.ledger,
		from:   b.slotIndex,
		to:     slot,
		done:   b.catchUpDone,
	}.catchUp()
}

func (b *ballotProtocol) finishCatchUp() {
	for _, slot := range b.persistLater {
		b.ledger.PersistSlot(slot)
	}

	b.slotIndex = b.persistLater[len(b.persistLater)-1].Index + 1
	b.persistLater = nil
}

func (b *ballotProtocol) recomputeCurrentBallotValue() {
	if b.highestConfirmedPrepared != nil {
		b.currentBallot.value = b.highestConfirmedPrepared.value
		return
	}

	if b.lastCandidate != nil {
		b.currentBallot.value = b.lastCandidate
		return
	}

	if b.highestAcceptedPrepared != nil {
		b.currentBallot.value = b.highestAcceptedPrepared.value
		return
	}
}

func (b *ballotProtocol) updateBallotCounter(nodeID string, c uint32) {
	current := b.counters[nodeID]
	if current >= c {
		return
	}

	b.counters[nodeID] = c
	lowest := b.counters.lowestNonBlocking(b.currentBallot.counter, b.quorumSlices)
	if lowest > b.currentBallot.counter {
		b.updateCurrentBallotCounter(lowest)
	}

	if !b.counters.hasLesserThan(b.currentBallot.counter) {
		b.newTimer()
	}
}

func (b *ballotProtocol) newTimer() {
	b.timer.Stop()
	b.timer = time.NewTimer(time.Duration(b.currentBallot.counter+1) * time.Second)
}
