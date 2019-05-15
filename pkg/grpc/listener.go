package grpc

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/xdarksome/scp"
	"github.com/xdarksome/scp/key"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Listener struct {
	nodes    map[string]*remoteNode
	messages chan *scp.Message
}

func NewListener(addresses map[string]key.Public) *Listener {
	l := &Listener{
		nodes:    map[string]*remoteNode{},
		messages: make(chan *scp.Message, 1000000),
	}

	for addr, id := range addresses {
		l.nodes[id.String()] = newRemoteNode(id, addr, l.messages)
	}

	return l
}

func (l *Listener) listenTo(n *remoteNode) {
	for {
		if err := n.listen(); err != nil {
			logrus.WithError(err).Error("failed to listen to %s (%s)", n.ID.String(), n.address)
			time.Sleep(5 * time.Second)
		}
	}
}

func (l *Listener) Run() {
	for i := range l.nodes {
		go l.listenTo(l.nodes[i])
	}
}

func (l *Listener) Stop() {
	for _, node := range l.nodes {
		node.stop <- struct{}{}
	}
}

type remoteNode struct {
	ID      key.Public
	address string

	stop     chan struct{}
	messages chan *scp.Message
}

func newRemoteNode(id key.Public, address string, messages chan *scp.Message) *remoteNode {
	return &remoteNode{
		ID:       id,
		address:  address,
		stop:     make(chan struct{}),
		messages: messages,
	}
}

func (r *remoteNode) newMessage(m *Message) {
	r.messages <- &scp.Message{
		Type:      scp.MessageType(m.Type),
		SlotIndex: m.SlotIndex,
		NodeID:    r.ID.String(),
		Counter:   m.Counter,
		Value:     m.Value,
	}
}

func (r *remoteNode) listen() error {
	conn, err := grpc.Dial(r.address, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to dial tcp")
	}
	defer conn.Close()

	client := NewNodeClient(conn)
	streamChannel, err := client.StreamMessages(context.Background(), &StreamMessagesRequest{})
	if err != nil {
		return errors.Wrap(err, "failed to get messageChannel")
	}

	for {
		select {
		case <-r.stop:
			return nil
		default:
			m, err := streamChannel.Recv()
			if err != nil {
				return errors.Wrap(err, "failed to receive message")
			}
			r.newMessage(m)
		}
	}
}
