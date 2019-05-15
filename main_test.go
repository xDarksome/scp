package scp_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/xdarksome/scp/pkg/grpc"

	"github.com/sirupsen/logrus"
	"github.com/xdarksome/scp"
	"github.com/xdarksome/scp/key"
)

type validator struct{}
type badValidator struct{}

func (v validator) ValidateValue(scp.Value) bool {
	return true
}

func (v validator) CombineValues(values ...scp.Value) (composite scp.Value) {
	for _, v := range values {
		composite = append(composite, v...)
	}

	sort.Slice(composite, func(i, j int) bool {
		return composite[i] < composite[j]
	})

	return composite
}

func (v validator) PersistSlot(slot scp.Slot) {}

func (v badValidator) ValidateValue(segment scp.Value) bool {
	return false
}

func (v badValidator) CombineValues(values ...scp.Value) (composite scp.Value) {
	return
}

func (v badValidator) PersistSlot(slot scp.Slot) {}

func Test(t *testing.T) {
	node1key, _ := key.Generate()
	node1key.Alias = "alice"
	logrus.Print(node1key.String())
	node2key, _ := key.Generate()
	node2key.Alias = "bob"
	logrus.Print(node2key.String())
	node3key, _ := key.Generate()
	node3key.Alias = "charley"
	logrus.Print(node3key.String())
	node4key, _ := key.Generate()
	node4key.Alias = "dave"
	logrus.Print(node4key.String())

	streaming1 := grpc.NewStreaming(7001, node1key, map[string]key.Public{
		"localhost:7002": node2key,
		"localhost:7003": node3key,
		"localhost:7004": node4key,
	})
	go streaming1.Run()
	node1 := scp.NewNode(scp.Config{
		Key:     node1key,
		App:     validator{},
		Network: streaming1,
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 3,
				Validators: map[string]struct{}{
					node2key.String(): {},
					node3key.String(): {},
					node4key.String(): {},
				},
			},
		},
	})

	streaming2 := grpc.NewStreaming(7002, node2key, map[string]key.Public{
		"localhost:7001": node1key,
		"localhost:7003": node3key,
		"localhost:7004": node4key,
	})
	go streaming2.Run()
	node2 := scp.NewNode(scp.Config{
		Key:     node2key,
		App:     validator{},
		Network: streaming2,
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 3,
				Validators: map[string]struct{}{
					node1key.String(): {},
					node3key.String(): {},
					node4key.String(): {},
				},
			},
		},
	})

	streaming3 := grpc.NewStreaming(7003, node3key, map[string]key.Public{
		"localhost:7002": node2key,
		"localhost:7001": node1key,
		"localhost:7004": node4key,
	})
	go streaming3.Run()
	node3 := scp.NewNode(scp.Config{
		Key:     node3key,
		App:     validator{},
		Network: streaming3,
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 3,
				Validators: map[string]struct{}{
					node2key.String(): {},
					node1key.String(): {},
					node4key.String(): {},
				},
			},
		},
	})

	streaming4 := grpc.NewStreaming(7004, node4key, map[string]key.Public{
		"localhost:7002": node2key,
		"localhost:7003": node3key,
		"localhost:7001": node1key,
	})
	go streaming4.Run()
	node4 := scp.NewNode(scp.Config{
		Key:     node4key,
		App:     validator{},
		Network: streaming4,
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 3,
				Validators: map[string]struct{}{
					node2key.String(): {},
					node3key.String(): {},
					node1key.String(): {},
				},
			},
		},
	})

	go node1.Run()
	go node2.Run()
	go node3.Run()
	go node4.Run()

	time.Sleep(2 * time.Second)

	go func() {
		for {
			node1.Propose([]byte(fmt.Sprint(rand.Int())))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	go func() {
		for {
			node1.Propose([]byte(fmt.Sprint(rand.Int())))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		node3.Propose([]byte(fmt.Sprint(rand.Int())))
		time.Sleep(500 * time.Millisecond)
	}
}
