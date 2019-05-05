package scp

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xdarksome/scp/pkg/key"
	"github.com/xdarksome/scp/pkg/network"
	"google.golang.org/grpc"
)

type listener struct {
	node    *Node
	nodeID  key.Public
	address string

	stop   chan struct{}
	output chan *network.Message
}

func newListener(n *Node, nodeID key.Public, address string, output chan *network.Message) *listener {
	return &listener{
		node:    n,
		nodeID:  nodeID,
		address: address,
		stop:    make(chan struct{}),
		output:  output,
	}
}

func (l *listener) Run() {
	for {
		if err := l.listen(); err != nil {
			logrus.WithError(err).Error("failed to listen")
			time.Sleep(5 * time.Second)
		}
	}
}

func (l *listener) listen() error {
	conn, err := grpc.Dial(l.address, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to dial tcp")
	}
	defer conn.Close()

	client := network.NewTenvisClient(conn)
	stream, err := client.Streaming(context.Background(), &network.StreamingRequest{
		Test: l.node.ID.String(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to get stream")
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
			l.output <- m
		}
	}
}
