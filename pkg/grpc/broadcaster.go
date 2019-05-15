package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xdarksome/scp"
	"github.com/xdarksome/scp/key"
	"google.golang.org/grpc"
)

type messageChannel struct {
	messages chan *Message
	done     chan struct{}
}

func (s *messageChannel) Done() {
	s.done <- struct{}{}
}

func newMessageChannel() messageChannel {
	return messageChannel{
		make(chan *Message, 100),
		make(chan struct{}),
	}
}

type Broadcaster struct {
	nodeID key.Public
	port   uint

	input  chan *scp.Message
	output map[chan struct{}]messageChannel
	queue  chan messageChannel
}

func NewBroadcaster(port uint, key key.Public) *Broadcaster {
	return &Broadcaster{
		nodeID: key,
		port:   port,
		input:  make(chan *scp.Message, 1000000),
		output: make(map[chan struct{}]messageChannel),
		queue:  make(chan messageChannel),
	}
}

func (b *Broadcaster) broadcast(m *scp.Message) {
	message := &Message{
		Type:      Message_Type(m.Type),
		SlotIndex: m.SlotIndex,
		Counter:   m.Counter,
		Value:     m.Value,
	}

	for done, ch := range b.output {
		select {
		case <-done:
			delete(b.output, done)
		default:
			ch.messages <- message
		}
	}
}

func (b *Broadcaster) Run() {
	go b.Serve()

	for {
		select {
		case m := <-b.input:
			b.broadcast(m)
		case ch := <-b.queue:
			b.output[ch.done] = ch
		}
	}
}

func (b *Broadcaster) Serve() {
	for {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", b.port))
		if err != nil {
			logrus.WithError(err).Error("failed to start tcp Listener")
			time.Sleep(time.Second)
			continue
		}
		server := grpc.NewServer()
		RegisterNodeServer(server, b)
		if err := server.Serve(l); err != nil {
			logrus.WithError(err).Error("grpc server failed")
			time.Sleep(time.Second)
			continue
		}
	}
}

func (b *Broadcaster) newStream() messageChannel {
	s := newMessageChannel()
	b.queue <- s
	return s
}

func (b *Broadcaster) StreamMessages(req *StreamMessagesRequest, stream Node_StreamMessagesServer) error {
	ch := b.newStream()

	for {
		m := <-ch.messages
		if err := stream.Send(m); err != nil {
			ch.Done()
			return err
		}
	}
}

func (b *Broadcaster) GetSlot(ctx context.Context, req *GetSlotRequest) (*Slot, error) {
	return nil, nil
}
