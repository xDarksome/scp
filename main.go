package scp

import (
	"fmt"

	"github.com/xdarksome/scp/internal/btree"
	"github.com/xdarksome/scp/internal/quorum"
	"github.com/xdarksome/scp/pkg/key"
	"github.com/xdarksome/scp/pkg/network"
)

type QuorumSlice quorum.Slice

type Validator interface {
	ValidateSegment(*network.Value) bool
}

type NodeConfig struct {
	Key         key.Public
	Validator   Validator
	QuorumSlice QuorumSlice
	Addresses   map[string]key.Public
	Port        uint
}

type Node struct {
	ID key.Public
	Validator
	quorumSlice quorum.Slice

	listeners map[string]*listener
	input     chan *network.Message

	block      uint64
	nominates  *btree.Agreements
	candidates []*network.Value

	output      chan *network.Message
	broadcaster broadcaster
}

func NewNode(cfg NodeConfig) *Node {
	cfg.QuorumSlice.Validators = append(cfg.QuorumSlice.Validators, cfg.Key.String())
	n := &Node{
		ID:          cfg.Key,
		Validator:   cfg.Validator,
		quorumSlice: quorum.Slice(cfg.QuorumSlice),
		listeners:   map[string]*listener{},
		input:       make(chan *network.Message, 1000000),
		nominates:   btree.NewAgreements(100),
		output:      make(chan *network.Message, 1000000),
	}

	for addr, id := range cfg.Addresses {
		n.listeners[id.String()] = newListener(n, id, addr, n.input)
	}

	n.broadcaster = newBroadcaster(cfg.Key, cfg.Port, n.output)

	return n
}

func (n *Node) ConsiderSegment(segment *network.Value) {
	if !n.ValidateSegment(segment) {
		return
	}

	agreement, _ := n.nominates.GetOrInsert(quorum.AgreementOn(segment), n.ID.String(), n.quorumSlice)
	n.proposeNominate(agreement)
}

func (n *Node) proposeNominate(nominate *quorum.Agreement) {
	fmt.Println(n.ID.String(), "proposed", nominate.Subject)
	n.output <- &network.Message{
		Nominate: &network.Nominate{
			Voted: []*network.Value{nominate.Subject},
		},
	}
	nominate.VotedBy(n.ID.String())

	if !nominate.Accepted && nominate.Vote.ReachedQuorumThreshold() {
		n.acceptNominate(nominate)
	}
}

func (n *Node) acceptNominate(nominate *quorum.Agreement) {
	fmt.Println(n.ID.String(), "accepted", nominate.Subject)
	n.output <- &network.Message{
		Nominate: &network.Nominate{
			Accepted: []*network.Value{nominate.Subject},
		},
	}
	nominate.AcceptedBy(n.ID.String())

	if !nominate.Confirmed && nominate.Accept.ReachedQuorumThreshold() {
		n.confirmNominate(nominate)
	}

	nominate.Accepted = true
}

func (n *Node) confirmNominate(nominate *quorum.Agreement) {
	n.candidates = append(n.candidates, nominate.Subject)
	nominate.Confirmed = true
}

func (n *Node) Run() {
	for i := range n.listeners {
		go n.listeners[i].Run()
	}
	go n.broadcaster.Run()

	for m := range n.input {
		if m.Nominate.Voted != nil {
			n.nominateProposed(m.NodeID, m.Nominate.Voted)
		}
		if m.Nominate.Accepted != nil {
			n.nominateAccepted(m.NodeID, m.Nominate.Accepted)
		}
	}
}

func (n *Node) nominateProposed(nodeID string, segments []*network.Value) {
	for i := range segments {
		s := segments[i]
		nominate, inserted := n.nominates.GetOrInsert(quorum.AgreementOn(s), n.ID.String(), n.quorumSlice)
		nominate.VotedBy(nodeID)

		if !inserted {
			if !nominate.Accepted && nominate.Vote.ReachedQuorumThreshold() {
				n.acceptNominate(nominate)
			}
			continue
		}

		if !n.ValidateSegment(s) {
			continue
		}

		n.proposeNominate(nominate)
	}
}

func (n *Node) nominateAccepted(nodeID string, segments []*network.Value) {
	for i := range segments {
		s := segments[i]

		nominate, _ := n.nominates.GetOrInsert(quorum.AgreementOn(s), n.ID.String(), n.quorumSlice)

		nominate.AcceptedBy(nodeID)
		if !nominate.Accepted && nominate.Accept.ReachedBlockingThreshold() {
			n.acceptNominate(nominate)
			continue
		}

		if !nominate.Confirmed && nominate.Accept.ReachedQuorumThreshold() {
			n.confirmNominate(nominate)
		}
	}
}
