package scp

import "github.com/sirupsen/logrus"

type Value []byte

type Slot struct {
	Index uint64
	Value
}

type Config struct {
	NodeID       string
	CurrentSlot  uint64
	Validator    Validator
	Combiner     Combiner
	Ledger       Ledger
	QuorumSlices []*QuorumSlice
}

type Consensus struct {
	nominationProtocol
	ballotProtocol

	outputMessages chan *Message
}

func New(cfg Config) *Consensus {
	outputMessages := make(chan *Message, 1000000)
	link := make(chan protocolMessage)
	candidates := make(chan Value, 1000)

	nominationProtocol := nominationProtocol{
		slotIndex:      cfg.CurrentSlot,
		id:             cfg.NodeID,
		validator:      cfg.Validator,
		combiner:       cfg.Combiner,
		quorumSlices:   cfg.QuorumSlices,
		proposals:      newSuspendableValueCh(),
		inputMessages:  make(chan *Message, 1000000),
		outputMessages: outputMessages,
		ballotProtocol: link,
	}
	nominationProtocol.init(cfg.CurrentSlot, candidates)

	ballotProtocol := ballotProtocol{
		slotIndex:          cfg.CurrentSlot,
		id:                 cfg.NodeID,
		ledger:             cfg.Ledger,
		quorumSlices:       cfg.QuorumSlices,
		inputMessages:      make(chan *Message, 1000000),
		outputMessages:     outputMessages,
		candidates:         candidates,
		nominationProtocol: link,
	}
	ballotProtocol.init(cfg.CurrentSlot)

	return &Consensus{
		nominationProtocol: nominationProtocol,
		ballotProtocol:     ballotProtocol,
		outputMessages:     outputMessages,
	}
}

func (c *Consensus) Run() {
	go c.nominationProtocol.run()
	go c.ballotProtocol.run()
}

func (c *Consensus) Propose(v Value) {
	c.nominationProtocol.proposals.in <- v
}

func (c *Consensus) InputMessage(m *Message) {
	switch m.Type {
	case VoteNominate, AcceptNominate:
		c.nominationProtocol.inputMessages <- m
	case VotePrepare, AcceptPrepare, VoteCommit, AcceptCommit:
		c.ballotProtocol.inputMessages <- m
	default:
		logrus.Errorf("unknown message type: %d", m.Type)
	}
}

func (c *Consensus) OutputMessage() *Message {
	return <-c.outputMessages
}

type suspendableValueCh struct {
	in  chan Value
	out chan Value
}

func newSuspendableValueCh() suspendableValueCh {
	ch := suspendableValueCh{
		in: make(chan Value),
	}
	ch.out = ch.in

	return ch
}

func (p *suspendableValueCh) resume() {
	p.out = p.in
}

func (p *suspendableValueCh) suspend() {
	p.out = nil
}
