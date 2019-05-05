package scp

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/xdarksome/scp/pkg/key"

	"github.com/xdarksome/scp/pkg/network"
)

type validator struct{}
type badValidator struct{}

func (v validator) ValidateSegment(segment *network.Value) bool {
	return true
}

func (v badValidator) ValidateSegment(segment *network.Value) bool {
	return false
}

func Test(t *testing.T) {
	node1key, _ := key.Generate()
	logrus.Print(node1key.String())
	node2key, _ := key.Generate()
	logrus.Print(node2key.String())
	node3key, _ := key.Generate()
	logrus.Print(node3key.String())
	node4key, _ := key.Generate()
	logrus.Print(node4key.String())

	node1 := NewNode(NodeConfig{
		Key:       node1key,
		Validator: validator{},
		QuorumSlice: QuorumSlice{
			Threshold:  3,
			Validators: []string{node2key.String(), node3key.String(), node4key.String()},
		},
		Addresses: map[string]key.Public{
			"localhost:7002": node2key,
			"localhost:7003": node3key,
			"localhost:7004": node4key,
		},
		Port: 7001,
	})

	node2 := NewNode(NodeConfig{
		Key:       node2key,
		Validator: validator{},
		QuorumSlice: QuorumSlice{
			Threshold:  3,
			Validators: []string{node1key.String(), node3key.String(), node4key.String()},
		},
		Addresses: map[string]key.Public{
			"localhost:7001": node1key,
			"localhost:7003": node3key,
			"localhost:7004": node4key,
		},
		Port: 7002,
	})

	node3 := NewNode(NodeConfig{
		Key:       node3key,
		Validator: validator{},
		QuorumSlice: QuorumSlice{
			Threshold:  3,
			Validators: []string{node2key.String(), node1key.String(), node4key.String()},
		},
		Addresses: map[string]key.Public{
			"localhost:7002": node2key,
			"localhost:7001": node1key,
			"localhost:7004": node4key,
		},
		Port: 7003,
	})

	node4 := NewNode(NodeConfig{
		Key:       node4key,
		Validator: badValidator{},
		QuorumSlice: QuorumSlice{
			Threshold:  3,
			Validators: []string{node2key.String(), node3key.String(), node1key.String()},
		},
		Addresses: map[string]key.Public{
			"localhost:7002": node2key,
			"localhost:7003": node3key,
			"localhost:7001": node1key,
		},
		Port: 7004,
	})

	go node1.Run()
	time.Sleep(time.Second)
	_ = node2
	go node2.Run()
	time.Sleep(time.Second)
	_ = node3
	go node3.Run()
	time.Sleep(time.Second)
	_ = node4
	go node4.Run()
	time.Sleep(time.Second)

	for i := 1; ; i++ {
		node1.ConsiderSegment(&network.Value{
			Data: []byte(fmt.Sprint(i)),
		})
		time.Sleep(5 * time.Second)
	}
}
