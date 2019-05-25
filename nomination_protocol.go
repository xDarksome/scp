package scp

type Validator interface {
	ValidateValue(Value) bool
}

type Combiner interface {
	CombineValues(...Value) Value
}

type nominationProtocol struct {
	slotIndex uint64
	id        string

	validator Validator
	combiner  Combiner

	quorumSlices
	nominates

	proposals  suspendableValueCh
	candidates chan Value

	inputMessages  chan *Message
	outputMessages chan *Message

	ballotProtocol chan protocolMessage
	lastCandidate  Value
}

func (n *nominationProtocol) run() {
	for {
		select {
		case p := <-n.proposals.out:
			n.considerProposal(p)
		case m := <-n.ballotProtocol:
			n.reinit(m)
		case m := <-n.inputMessages:
			n.receive(m)
		}
	}
}

func (n *nominationProtocol) considerProposal(v Value) {
	if !n.validator.ValidateValue(v) {
		return
	}

	if nominate := n.nominates.Find(v); nominate == nil {
		nominate = n.nominates.Create(v, n.quorumSlices)
		n.voteNominate(nominate)
	}
}

func (n *nominationProtocol) receive(m *Message) {
	if m.SlotIndex != n.slotIndex {
		return
	}

	switch m.Type {
	case VoteNominate:
		n.nominateVoted(m)
	case AcceptNominate:
		n.nominateAccepted(m)
	}
}

func (n *nominationProtocol) broadcast(m *Message) {
	if m.SlotIndex != n.slotIndex {
		return
	}

	m.NodeID = n.id
	n.outputMessages <- m
}

func (n *nominationProtocol) init(slotIndex uint64, candidatesCh chan Value) {
	n.slotIndex = slotIndex
	n.nominates = newNominates()
	n.candidates = candidatesCh
}

func (n *nominationProtocol) reinit(m protocolMessage) {
	n.init(m.slotIndex, m.candidates)
	n.lastCandidate = nil
	n.proposals.resume()
}

func (n *nominationProtocol) nominateVoted(m *Message) {
	nominate := n.nominates.Find(m.Value)
	if nominate == nil {
		if !n.validator.ValidateValue(m.Value) {
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

func (n *nominationProtocol) nominateAccepted(m *Message) {
	//fmt.Println(n.id, "see", m.NodeID, "has accepted nominate", m.SlotIndex, string(m.Value))
	nominate := n.nominates.FindOrCreate(m.Value, n.quorumSlices)
	nominate.acceptedBy(m.NodeID)

	if !nominate.accepted && nominate.accepts.reachedBlockingThreshold(n.quorumSlices) {
		n.acceptNominate(nominate)
	}

	if !nominate.confirmed && nominate.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmNominate(nominate)
	}
}

func (n *nominationProtocol) voteNominate(nominate *nominate) {
	//fmt.Println(n.id, "votes for nominate", string(nominate.value))
	n.broadcast(&Message{
		Type:      VoteNominate,
		SlotIndex: n.slotIndex,
		Value:     nominate.value,
	})
	nominate.selfVote()

	if !nominate.accepted && nominate.votes.reachedQuorumThreshold(n.quorumSlices) {
		n.acceptNominate(nominate)
	}
}

func (n *nominationProtocol) acceptNominate(nominate *nominate) {
	//fmt.Println(n.id, "accepts nominate", string(nominate.value))
	n.broadcast(&Message{
		Type:      AcceptNominate,
		SlotIndex: n.slotIndex,
		Value:     nominate.value,
	})
	nominate.selfAccept()

	if !nominate.confirmed && nominate.accepts.reachedQuorumThreshold(n.quorumSlices) {
		n.confirmNominate(nominate)
	}
}

func (n *nominationProtocol) confirmNominate(nominate *nominate) {
	//fmt.Println(n.id, "confirms nominate", string(nominate.value))
	n.proposals.suspend()
	nominate.selfConfirm()

	n.newCandidate(nominate.value)
}

func (n *nominationProtocol) newCandidate(c Value) {
	n.lastCandidate = n.combiner.CombineValues(n.lastCandidate, c)
	n.candidates <- n.lastCandidate
}
