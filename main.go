package scp

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xdarksome/scp/key"
)

type Value []byte

type Slot struct {
	Index uint64
	Value
}

type App interface {
	Validator
	CombineValues(...Value) Value
	PersistSlot(Slot)
}

type Config struct {
	App          App
	Network      Network
	Key          key.Public
	QuorumSlices []*QuorumSlice
}

type Node struct {
	App
	slotIndex uint64

	quorumSlices
	network       Network
	inputMessages chan *Message
	proposals
	nominationProtocol *nominationProtocol
	candidate          Value
	ballotProtocol     *ballotProtocol
}

func NewNode(cfg Config) *Node {
	return &Node{
		App:           cfg.App,
		network:       cfg.Network,
		quorumSlices:  cfg.QuorumSlices,
		inputMessages: make(chan *Message, 1000000),
		proposals:     newProposals(),
	}
}

func (n *Node) Propose(v Value) {
	n.proposals.input <- v
}

func (n *Node) considerProposal(v Value) {
	if !n.ValidateValue(v) {
		return
	}
	n.nominationProtocol.proposals <- v
}

func (n *Node) newSlot() {
	n.slotIndex++
	n.candidate = nil
	n.proposals.start()

	n.nominationProtocol = newNominationProtocol(n.slotIndex, n.quorumSlices, n.network, n.App)
	go n.nominationProtocol.Run()

	n.ballotProtocol = newBallotProtocol(n.slotIndex, n.quorumSlices, n.network)
	go n.ballotProtocol.Run()
}

func (n *Node) fetchNetworkMessages() {
	var m *Message
	for {
		m = n.network.Receive()
		n.inputMessages <- m
	}
}

func (n *Node) receiveMessage(m *Message) {
	if m.SlotIndex < n.slotIndex {
		return
	}

	switch m.Type {
	case VoteNominate, AcceptNominate:
		n.nominationProtocol.input <- m
	case VotePrepare, AcceptPrepare, VoteCommit, AcceptCommit:
		n.ballotProtocol.input <- m
	default:
		logrus.Errorf("unknown message type: %d", m.Type)
	}
}

func (n *Node) sendMessage(m *Message) {
	if m.SlotIndex != n.slotIndex {
		return
	}

	n.network.Broadcast(m)
}

func (n *Node) Run() {
	go n.fetchNetworkMessages()

	n.newSlot()
	for {
		select {
		case v := <-n.proposals.output:
			n.considerProposal(v)
		case m := <-n.inputMessages:
			fmt.Println(n.network.ID(), "received", m)
			n.receiveMessage(m)
		case c := <-n.nominationProtocol.candidates:
			n.proposals.stop()
			n.candidate = n.CombineValues(n.candidate, c)
			n.ballotProtocol.candidates <- n.candidate
		case s := <-n.ballotProtocol.slots:
			n.nominationProtocol.Stop()
			n.ballotProtocol.Stop()
			fmt.Println("externalized", s.Index, s.Value)
			n.PersistSlot(s)
			n.newSlot()
		}
	}
}
