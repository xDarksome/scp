package scp

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type Value []byte

type Slot struct {
	Index uint64
	Value
}

type Ledger interface {
	PersistSlot(Slot)
}

type Validator interface {
	ValidateValue(Value) bool
}

type Combiner interface {
	CombineValues(...Value) Value
}

type Catchuper interface {
	GetSlot(slotIndex uint64, nodeID string) (Slot, error)
}

type Config struct {
	NodeID       string
	Validator    Validator
	Combiner     Combiner
	Ledger       Ledger
	QuorumSlices []*QuorumSlice
}

type Node struct {
	id string
	Validator
	Combiner
	Ledger

	quorumSlices
	messages   messages
	proposals  proposals
	nominates  nominates
	candidates candidates
	ballots    ballots
	slots      slots
}

func New(lastSlot uint64, cfg Config) *Node {
	return &Node{
		id:           cfg.NodeID,
		Validator:    cfg.Validator,
		Combiner:     cfg.Combiner,
		Ledger:       cfg.Ledger,
		quorumSlices: cfg.QuorumSlices,
		messages:     newMessages(),
		proposals:    newProposals(),
		slots:        newSlots(lastSlot),
	}
}

func (n *Node) Propose(v Value) {
	n.proposals.input <- v
}

func (n *Node) considerProposal(v Value) {
	if !n.ValidateValue(v) {
		return
	}

	if nominate := n.nominates.Find(v); nominate == nil {
		nominate = n.nominates.Create(v, n.quorumSlices)
		n.voteNominate(nominate)
	}
}

func (n *Node) newSlot() (done func()) {
	n.slots.currentIndex++
	n.proposals.open()

	n.nominates = newNominates()
	n.candidates = newCandidates()
	n.ballots = newBallots(n.slots.currentIndex)

	stopNominationProtocol := n.runNominationProtocol()
	stopBallotProtocol := n.runBallotProtocol()

	return func() {
		stopNominationProtocol()
		stopBallotProtocol()
	}
}

func (n *Node) InputMessage(m *Message) {
	n.messages.input <- m
}

func (n *Node) OutputMessage() *Message {
	return <-n.messages.output
}

func (n *Node) receiveMessage(m *Message) {
	if m.SlotIndex < n.slots.currentIndex {
		return
	}

	switch m.Type {
	case VoteNominate:
		n.messages.voteNominate <- m
	case AcceptNominate:
		n.messages.acceptNominate <- m
	case VotePrepare:
		n.messages.votePrepare <- m
	case AcceptPrepare:
		n.messages.acceptPrepare <- m
	case VoteCommit:
		n.messages.voteCommit <- m
	case AcceptCommit:
		n.messages.acceptCommit <- m
	default:
		logrus.Errorf("unknown message type: %d", m.Type)
	}
}

func (n *Node) broadcast(m *Message) {
	if m.SlotIndex != n.slots.currentIndex {
		return
	}

	m.NodeID = n.id
	n.messages.output <- m
}

func (n *Node) Run() {
	done := n.newSlot()
	for {
		select {
		case m := <-n.messages.input:
			n.receiveMessage(m)
		case s := <-n.slots.output:
			n.PersistSlot(s)
			fmt.Println("externalized", s)
			done()
			done = n.newSlot()
		}
	}
}

func (n *Node) runNominationProtocol() (stop func()) {
	stopCh := make(chan struct{})
	go func() {
		for {
			select {
			case p := <-n.proposals.output:
				n.considerProposal(p)
			case m := <-n.messages.voteNominate:
				if n.slots.currentIndex == m.SlotIndex {
					n.nominateVoted(m)
				}
			case m := <-n.messages.acceptNominate:
				if n.slots.currentIndex == m.SlotIndex {
					n.nominateAccepted(m)
				}
			case <-stopCh:
				return
			}
		}
	}()

	return func() {
		stopCh <- struct{}{}
	}
}

func (n *Node) runBallotProtocol() (stop func()) {
	stopCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-n.ballots.timer.C:
				n.updateCurrentBallotCounter(n.ballots.current.counter + 1)
			case c := <-n.candidates.queue:
				n.newCandidate(c)
			case m := <-n.messages.votePrepare:
				n.prepareVoted(m)
			case m := <-n.messages.acceptPrepare:
				n.prepareAccepted(m)
			case m := <-n.messages.voteCommit:
				n.commitVoted(m)
			case m := <-n.messages.acceptCommit:
				n.commitAccepted(m)
			case <-stopCh:
				return
			}
		}
	}()

	return func() {
		stopCh <- struct{}{}
	}
}

func (n *Node) nominateVoted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has voted nominate", m.Value)
	nominate := n.nominates.Find(m.Value)
	if nominate == nil {
		if !n.ValidateValue(m.Value) {
			//logrus.Warnf("invalid nominate from %s", m.NodeID)
			return
		}
		nominate = n.nominates.Create(m.Value, n.quorumSlices)
		n.voteNominate(nominate)
	}

	nominate.votedBy(m.NodeID)
	if !nominate.accepted && nominate.votes.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptNominate(nominate)
	}
}

func (n *Node) nominateAccepted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has accepted nominate", m.Value)
	nominate := n.nominates.FindOrCreate(m.Value, n.quorumSlices)
	nominate.acceptedBy(m.NodeID)

	if !nominate.accepted && nominate.accepts.reachedBlockingThreshold(n.quorumSlices) {
		n.acceptNominate(nominate)
	}

	if !nominate.confirmed && nominate.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmNominate(nominate)
	}
}

func (n *Node) prepareVoted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has voted prepare", m.Value)

	n.updateBallotCounter(m.NodeID, m.Counter)
	ballot := n.ballots.FindOrCreate(m.SlotIndex, m.Counter, m.Value, n.quorumSlices)
	ballot.prepare.votedBy(m.NodeID)

	if !ballot.prepare.accepted && ballot.prepare.votes.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptPrepare(ballot)
	}
}

func (n *Node) prepareAccepted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has accepted prepare", m.Value)
	n.updateBallotCounter(m.NodeID, m.Counter)
	ballot := n.ballots.FindOrCreate(m.SlotIndex, m.Counter, m.Value, n.quorumSlices)
	ballot.prepare.acceptedBy(m.NodeID)

	if !ballot.prepare.accepted && ballot.prepare.accepts.reachedBlockingThreshold(n.quorumSlices) {
		n.acceptPrepare(ballot)
	}

	if !ballot.prepare.confirmed && ballot.prepare.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmPrepare(ballot)
	}
}

func (n *Node) commitVoted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has voted commit", m.Value)
	n.updateBallotCounter(m.NodeID, m.Counter)
	ballot := n.ballots.FindOrCreate(m.SlotIndex, m.Counter, m.Value, n.quorumSlices)
	ballot.commit.votedBy(m.NodeID)

	if !ballot.commit.accepted && ballot.commit.votes.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptCommit(ballot)
	}
}

func (n *Node) commitAccepted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has accepted commit", m.Value)
	n.updateBallotCounter(m.NodeID, m.Counter)
	ballot := n.ballots.FindOrCreate(m.SlotIndex, m.Counter, m.Value, n.quorumSlices)
	ballot.commit.acceptedBy(m.NodeID)

	if !ballot.commit.accepted && ballot.commit.accepts.reachedBlockingThreshold(n.quorumSlices) {
		n.acceptCommit(ballot)
	}

	if !ballot.commit.confirmed && ballot.commit.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmCommit(ballot)
	}
}

func (n *Node) voteNominate(nominate *nominate) {
	//fmt.Println(n.id, "votes for nominate", nominate.value)
	n.broadcast(&Message{
		Type:      VoteNominate,
		SlotIndex: n.slots.currentIndex,
		Value:     nominate.value,
	})
	nominate.selfVote()

	if !nominate.accepted && nominate.votes.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptNominate(nominate)
	}
}

func (n *Node) acceptNominate(nominate *nominate) {
	//fmt.Println(n.id, "accepts nominate", nominate.value)
	n.broadcast(&Message{
		Type:      AcceptNominate,
		SlotIndex: n.slots.currentIndex,
		Value:     nominate.value,
	})
	nominate.selfAccept()

	if !nominate.confirmed && nominate.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmNominate(nominate)
	}
}

func (n *Node) confirmNominate(nominate *nominate) {
	//fmt.Println(n.id, "confirms nominate", nominate.value)
	n.candidates.queue <- nominate.value
	nominate.selfConfirm()
	n.proposals.close()
}

func (n *Node) votePrepare() {
	ballot := n.ballots.FindOrCreate(
		n.ballots.current.slotIndex,
		n.ballots.current.counter,
		n.ballots.current.value,
		n.quorumSlices)
	n.ballots.current = ballot
	//fmt.Println(n.id, "votes for prepare", ballot.value)

	n.broadcast(&Message{
		Type:      VotePrepare,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.prepare.selfVote()

	if !ballot.prepare.accepted && ballot.prepare.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptPrepare(ballot)
	}
}

func (n *Node) acceptPrepare(ballot *ballot) {
	//fmt.Println(n.id, "accepts prepare", ballot.value)
	n.broadcast(&Message{
		Type:      AcceptPrepare,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.prepare.selfAccept()

	if n.ballots.highestAcceptedPrepared == nil || n.ballots.highestAcceptedPrepared.less(ballot) {
		n.ballots.highestAcceptedPrepared = ballot
	}

	if !ballot.prepare.confirmed && ballot.prepare.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmPrepare(ballot)
	}
}

func (n *Node) confirmPrepare(ballot *ballot) {
	//fmt.Println(n.id, "confirms prepare", ballot.value)
	if n.ballots.highestConfirmedPrepared == nil || n.ballots.highestConfirmedPrepared.less(ballot) {
		n.ballots.highestConfirmedPrepared = ballot
	}

	ballot.prepare.selfConfirm()
	if !ballot.less(n.ballots.current) || ballot.compatible(n.ballots.current) {
		n.voteCommit(ballot)
	}
}

func (n *Node) voteCommit(ballot *ballot) {
	//fmt.Println(n.id, "votes for commit", ballot.value)
	n.broadcast(&Message{
		Type:      VoteCommit,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.commit.selfVote()

	if !ballot.commit.accepted && ballot.commit.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptCommit(ballot)
	}
}

func (n *Node) acceptCommit(ballot *ballot) {
	//fmt.Println(n.id, "accepts commit", ballot.value)
	n.broadcast(&Message{
		Type:      AcceptCommit,
		SlotIndex: ballot.slotIndex,
		Counter:   ballot.counter,
		Value:     ballot.value,
	})
	ballot.commit.selfAccept()

	if !ballot.commit.confirmed && ballot.commit.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmCommit(ballot)
	}
}

func (n *Node) confirmCommit(ballot *ballot) {
	//fmt.Println(n.id, "confirms commit", ballot.value)
	ballot.commit.selfConfirm()
	n.slots.output <- Slot{
		Index: ballot.slotIndex,
		Value: ballot.value,
	}
}

func (n *Node) updateCurrentBallotCounter(c uint32) {
	n.ballots.current.counter = c
	n.recomputeCurrentBallotValue()
	n.votePrepare()
}

func (n *Node) recomputeCurrentBallotValue() {
	if n.ballots.highestConfirmedPrepared != nil {
		n.ballots.current.value = n.ballots.highestConfirmedPrepared.value
		return
	}

	if n.candidates.current != nil {
		n.ballots.current.value = n.candidates.current
		return
	}

	if n.ballots.highestAcceptedPrepared != nil {
		n.ballots.current.value = n.ballots.highestAcceptedPrepared.value
		return
	}
}

func (n *Node) updateBallotCounter(nodeID string, c uint32) {
	current := n.ballots.counters[nodeID]
	if current >= c {
		return
	}

	n.ballots.counters[nodeID] = c
	lowest := n.ballots.counters.lowestNonBlocking(n.ballots.current.counter, n.quorumSlices)
	if lowest > n.ballots.current.counter {
		n.updateCurrentBallotCounter(lowest)
	}

	if !n.ballots.counters.hasLesserThan(n.ballots.current.counter) {
		n.newBallotTimer()
	}
}

func (n *Node) newBallotTimer() {
	n.ballots.timer.Stop()
	n.ballots.timer = time.NewTimer(time.Duration(n.ballots.current.counter+1) * time.Second)
}

func (n *Node) newCandidate(value Value) {
	first := n.candidates.current == nil
	n.candidates.current = value

	if first {
		n.ballots.current.value = value
		n.votePrepare()
	}
}
