package scp

type Network interface {
	Receive() *Message
	Broadcast(*Message)
	ID() string
}

type MessageType int

const (
	VoteNominate MessageType = iota + 1
	AcceptNominate

	VotePrepare
	AcceptPrepare

	VoteCommit
	AcceptCommit
)

type Message struct {
	Type MessageType

	SlotIndex uint64
	NodeID    string

	Counter uint32
	Value
}

type messages struct {
	input chan *Message

	voteNominate   chan *Message
	acceptNominate chan *Message
	votePrepare    chan *Message
	acceptPrepare  chan *Message
	voteCommit     chan *Message
	acceptCommit   chan *Message

	output chan *Message
}

func newMessages() messages {
	return messages{
		input:          make(chan *Message, 1000000),
		voteNominate:   make(chan *Message, 1000),
		acceptNominate: make(chan *Message, 1000),
		votePrepare:    make(chan *Message, 1000),
		acceptPrepare:  make(chan *Message, 1000),
		voteCommit:     make(chan *Message, 1000),
		acceptCommit:   make(chan *Message, 1000),
		output:         make(chan *Message, 1000000),
	}

}

/*type network struct {
	broadcaster         *broadcaster
	receivedNominations chan *nominateMessage
	nominatingListeners map[string]*nominatingListener
}

func newNetwork(port uint, s *QuorumSlices) *network {
	net := &network{
		broadcaster:         newBroadcaster(s.self, port),
		receivedNominations: make(chan *nominateMessage),
		nominatingListeners: make(map[string]*nominatingListener),
	}

	for _, slice := range s.slices {
		for addr, id := range slice.Validators {
			net.nominatingListeners[id.String()] = newNominatingListener(id, addr, net.receivedNominations)
		}

		for _, inner := range slice.InnerSlices {
			for addr, id := range inner.Validators {
				net.nominatingListeners[id.String()] = newNominatingListener(id, addr, net.receivedNominations)
			}
		}
	}

	return net
}

func (n *network) Run() {
	n.broadcaster.Run()
}

func (n *network) ListenNominatesFrom(nodeID string) {
	listener := n.nominatingListeners[nodeID]
	if !listener.active {
		go listener.Run()
	}
}

func (n *network) StopNominatingListeners() {
	for _, l := range n.nominatingListeners {
		l.Stop()
	}
}



type ballotMessage struct {
	nodeID string
	*scp.BallotMessage
}

type nominateMessage struct {
	nodeID string
	*scp.NominateMessage
}

func newNominateMessage(nodeID string, m *scp.NominateMessage) *nominateMessage {
	return &nominateMessage{
		nodeID:          nodeID,
		NominateMessage: m,
	}
}

type nominatingListener struct {
	active  bool
	nodeID  key.Public
	address string

	stop   chan struct{}
	output chan *nominateMessage
}

func newNominatingListener(nodeID key.Public, address string, output chan *nominateMessage) *nominatingListener {
	return &nominatingListener{
		nodeID:  nodeID,
		address: address,
		stop:    make(chan struct{}),
		output:  output,
	}
}

func (l *nominatingListener) Run() {
	l.active = true
	for {
		if err := l.listen(); err != nil {
			logrus.WithError(err).Error("failed to listen")
			time.Sleep(5 * time.Second)
		}
	}
}

func (l *nominatingListener) Stop() {
	l.active = false
	l.stop <- struct{}{}
}

func (l *nominatingListener) listen() error {
	conn, err := grpc.Dial(l.address, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to dial tcp")
	}
	defer conn.Close()

	client := scp.NewNodeClient(conn)
	stream, err := client.NominatingStream(context.Background(), &scp.NominatingStreamRequest{})
	if err != nil {
		return errors.Wrap(err, "failed to get messageChannel")
	}

	for {
		select {
		case <-l.stop:
			return nil
		default:
			m, err := stream.Recv()
			if err != nil {
				return errors.Wrap(err, "failed to receive message")
			}
			l.output <- newNominateMessage(l.nodeID.String(), m)
		}
	}
}
*/
