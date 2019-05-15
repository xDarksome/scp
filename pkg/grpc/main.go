package grpc

import (
	"github.com/xdarksome/scp"
	"github.com/xdarksome/scp/key"
)

type Streaming struct {
	Listener
	Broadcaster
}

func NewStreaming(port uint, key key.Public, addresses map[string]key.Public) *Streaming {
	return &Streaming{
		Listener:    *NewListener(addresses),
		Broadcaster: *NewBroadcaster(port, key),
	}
}

func (s *Streaming) Run() {
	go s.Listener.Run()
	s.Broadcaster.Run()
}

func (s *Streaming) Receive() *scp.Message {
	return <-s.Listener.messages
}

func (s *Streaming) Broadcast(m *scp.Message) {
	s.Broadcaster.input <- m
}

func (s *Streaming) ID() string {
	return s.Broadcaster.nodeID.String()
}
