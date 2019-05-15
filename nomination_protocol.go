package scp

import (
	"bytes"

	"github.com/google/btree"

	"github.com/sirupsen/logrus"
)

type Validator interface {
	ValidateValue(Value) bool
}

type nominate struct {
	slotIndex uint64
	value     Value
	*ratification
}

func (n *nominate) Less(than btree.Item) bool {
	nom := than.(*nominate)

	if n.slotIndex < nom.slotIndex {
		return true
	}

	if n.slotIndex > nom.slotIndex {
		return false
	}

	return bytes.Compare(n.value, nom.value) == -1
}

type nominationProtocol struct {
	Validator
	quorumSlices
	nominates

	slotIndex uint64
	input     chan *Message

	proposals  chan Value
	candidates chan Value

	network Network

	stop chan struct{}
}

func newNominationProtocol(slot uint64, slices quorumSlices, network Network, validator Validator) *nominationProtocol {
	return &nominationProtocol{
		slotIndex:    slot,
		Validator:    validator,
		quorumSlices: slices,
		proposals:    make(chan Value, 1000000),
		nominates:    newNominates(),
		input:        make(chan *Message, 1000000),
		candidates:   make(chan Value, 1000),
		stop:         make(chan struct{}),
		network:      network,
	}
}

func (p *nominationProtocol) Run() {
	for {
		select {
		case pr := <-p.proposals:
			p.considerProposal(pr)
		case m := <-p.input:
			p.receiveMessage(m)
		case <-p.stop:
			return
		}
	}
}

func (p *nominationProtocol) Stop() {
	p.stop <- struct{}{}
}

func (p *nominationProtocol) receiveMessage(m *Message) {
	n := &nominate{slotIndex: m.SlotIndex, value: m.Value}
	switch m.Type {
	case VoteNominate:
		p.nominateVoted(m.NodeID, n)
	case AcceptNominate:
		p.nominateAccepted(m.NodeID, n)
	}
}

func (p *nominationProtocol) considerProposal(value Value) {
	n := &nominate{slotIndex: p.slotIndex, value: value}
	nominate := p.nominates.Get(n)
	if nominate != nil {
		return
	}

	nominate = p.nominates.Insert(n, p.quorumSlices)
	p.voteNominate(nominate)
}

func (p *nominationProtocol) nominateVoted(nodeID string, n *nominate) {
	//fmt.Println(p.network.ID(), "see", nodeID, "has voted nominate", n.value)
	nominate := p.nominates.Get(n)
	if nominate == nil {
		if !p.ValidateValue(n.value) {
			logrus.Warnf("invalid nominate from %s", nodeID)
			return
		}
		nominate = p.nominates.Insert(n, p.quorumSlices)
		p.voteNominate(nominate)
	}

	nominate.votedBy(nodeID)
	if !nominate.accepted && nominate.votes.reachedQuorumThreshold(p.quorumSlices) {
		p.acceptNominate(nominate)
	}
}

func (p *nominationProtocol) nominateAccepted(nodeID string, nominate *nominate) {
	//fmt.Println(p.network.ID(), "see", nodeID, "has accepted nominate", nominate.value)
	nominate = p.nominates.GetOrInsert(nominate, p.quorumSlices)
	nominate.acceptedBy(nodeID)

	if !nominate.accepted && nominate.accepts.reachedBlockingThreshold(p.quorumSlices) {
		p.acceptNominate(nominate)
	}

	if !nominate.confirmed && nominate.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.confirmNominate(nominate)
	}
}

func (p *nominationProtocol) voteNominate(nominate *nominate) {
	//fmt.Println(p.network.ID(), "votes for nominate", nominate.value)
	if nominate.slotIndex == p.slotIndex {
		p.network.Broadcast(&Message{
			Type:      VoteNominate,
			SlotIndex: nominate.slotIndex,
			Value:     nominate.value,
		})
	}
	nominate.selfVote()

	if !nominate.accepted && nominate.votes.reachedQuorumThreshold(p.quorumSlices) {
		p.acceptNominate(nominate)
	}
}

func (p *nominationProtocol) acceptNominate(nominate *nominate) {
	//fmt.Println(p.network.ID(), "accepts nominate", nominate.value)
	if nominate.slotIndex == p.slotIndex {
		p.network.Broadcast(&Message{
			Type:      AcceptNominate,
			SlotIndex: nominate.slotIndex,
			Value:     nominate.value,
		})
	}
	nominate.selfAccept()

	if !nominate.confirmed && nominate.accepts.reachedQuorumThreshold(p.quorumSlices) {
		p.confirmNominate(nominate)
	}
}

func (p *nominationProtocol) confirmNominate(nominate *nominate) {
	//fmt.Println(p.network.ID(), "confirms nominate", nominate.value)
	p.candidates <- nominate.value
	nominate.selfConfirm()
}

/*func (n *nominationProtocol) newRound() {
	n.round++
	n.newRoundTimer.Reset(time.Duration(n.round) * 500 * time.Millisecond)

	node := n.findHighestPriority()
	if node != n.nodeID.String() {
		n.network.ListenNominatesFrom(node)
		return
	}

	if n.localNominating {
		return
	}

	n.nominateVoted(n.nodeID.String(), n.pendingProposals...)
	n.localNominating = true
}*/

/*func (n *nominationProtocol) neighbors() (res []string) {
	Gi := scp.Gi{
		slotIndex:   n.slotIndex,
		Constant:    1,
		RoundNumber: n.round,
	}

	for _, nodeID := range n.quorumSlices.NodesList() {
		b := big.NewInt(1)
		b = b.Lsh(b, 256)
		nom, d := n.quorumSlices.NodeWeight(nodeID)
		b = b.Mul(b, big.NewInt(nom))
		b = b.Div(b, big.NewInt(d))

		Gi.NodeID = nodeID
		if gi(&Gi).Cmp(b) < 0 {
			res = append(res, nodeID)
		}
	}

	res = append(res, n.nodeID.String())
	return res
}

func (n *nominationProtocol) priority(nodeID string) *big.Int {
	return gi(&scp.Gi{
		slotIndex:   n.slotIndex,
		Constant:    2,
		RoundNumber: n.round,
		NodeID:      nodeID,
	})
}

func (n *nominationProtocol) findHighestPriority() (nodeID string) {
	highest := big.NewInt(0)
	for _, id := range n.neighbors() {
		priority := n.priority(id)
		if priority.Cmp(highest) > 0 {
			highest = priority
			nodeID = id
		}
	}

	return nodeID
}

func gi(gi *scp.Gi) *big.Int {
	b, err := proto.Marshal(gi)
	if err != nil {
		logrus.WithError(err).Panic("failed to marshal gi")
	}

	hash := sha256.Sum256(b)
	return new(big.Int).SetBytes(hash[:])
}
*/
