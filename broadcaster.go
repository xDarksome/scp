package scp

import (
	"fmt"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/xdarksome/scp/pkg/key"
	"github.com/xdarksome/scp/pkg/network"
	"google.golang.org/grpc"
)

const optimalMessageSize = 1 << 15

type stream struct {
	messages chan *network.Message
	done     chan struct{}
}

func newStream() stream {
	return stream{
		make(chan *network.Message, 100),
		make(chan struct{}),
	}
}

type broadcaster struct {
	nodeID  key.Public
	port    uint
	input   chan *network.Message
	streams map[chan struct{}]stream
	queue   chan stream
}

func newBroadcaster(nodeID key.Public, port uint, input chan *network.Message) broadcaster {
	return broadcaster{
		nodeID:  nodeID,
		input:   input,
		port:    port,
		streams: make(map[chan struct{}]stream),
		queue:   make(chan stream),
	}
}

func (b *broadcaster) Run() {
	go func() {
		for {
			l, err := net.Listen("tcp", fmt.Sprintf(":%d", b.port))
			if err != nil {
				logrus.WithError(err).Error("failed to start tcp listener")
				time.Sleep(time.Second)
				continue
			}
			server := grpc.NewServer()
			network.RegisterTenvisServer(server, b)
			if err := server.Serve(l); err != nil {
				logrus.WithError(err).Error("grpc server failed")
				time.Sleep(time.Second)
				continue
			}
		}
	}()

	for {
		select {
		case newStream := <-b.queue:
			b.streams[newStream.done] = newStream
		case m := <-b.input:
			for done, stream := range b.streams {
				select {
				case <-done:
					delete(b.streams, done)
				default:
					m.NodeID = b.nodeID.String()
					stream.messages <- m
				}
			}
		}
	}
}

func (b *broadcaster) newStream() stream {
	s := newStream()
	b.queue <- s
	return s
}

func (b *broadcaster) Streaming(req *network.StreamingRequest, stream network.Tenvis_StreamingServer) error {
	s := b.newStream()
	tick := time.Tick(100 * time.Millisecond)
	var buffer *network.Message
	send := func() error {
		if err := stream.Send(buffer); err != nil {
			logrus.WithError(err).Error("failed to send message")
			return err
		}
		stream.Context().Done()
		//fmt.Println("send", buffer)

		buffer = nil
		return nil
	}

	var err error
loop:
	for {
		select {
		case <-tick:
			if buffer != nil {
				if err = send(); err != nil {
					break loop
				}
			}
		case message := <-s.messages:
			if buffer == nil {
				buffer = message
				break
			}
			proto.Merge(buffer, message)
			if proto.Size(buffer) >= optimalMessageSize {
				if err = send(); err != nil {
					break loop
				}
			}
		}
	}

	s.done <- struct{}{}
	return err
}
