package scp_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/xdarksome/scp"
)

type app struct{}
type badApp struct{}

func (v app) ValidateValue(scp.Value) bool {
	return true
}

func (v app) CombineValues(values ...scp.Value) (composite scp.Value) {
	for _, v := range values {
		composite = append(composite, v...)
	}

	sort.Slice(composite, func(i, j int) bool {
		return composite[i] < composite[j]
	})

	return composite
}

func (v app) PersistSlot(slot scp.Slot) {}

func (v badApp) ValidateValue(segment scp.Value) bool {
	return false
}

func (v badApp) CombineValues(values ...scp.Value) (composite scp.Value) {
	return
}

func (v badApp) PersistSlot(slot scp.Slot) {}

func Test(t *testing.T) {
	node1key := "alice"
	node2key := "bob"
	node3key := "charley"
	node4key := "dave"

	node1 := scp.New(scp.Config{
		NodeID:      node1key,
		CurrentSlot: 10,
		Validator:   app{},
		Combiner:    app{},
		Ledger:      app{},
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 2,
				Validators: map[string]struct{}{
					node2key: {},
					node3key: {},
					node4key: {},
				},
			},
		},
	})

	node2 := scp.New(scp.Config{
		NodeID:      node2key,
		CurrentSlot: 10,
		Validator:   app{},
		Combiner:    app{},
		Ledger:      app{},
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 2,
				Validators: map[string]struct{}{
					node1key: {},
					node3key: {},
					node4key: {},
				},
			},
		},
	})

	node3 := scp.New(scp.Config{
		NodeID:      node3key,
		CurrentSlot: 10,
		Validator:   app{},
		Combiner:    app{},
		Ledger:      app{},
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 2,
				Validators: map[string]struct{}{
					node2key: {},
					node1key: {},
					node4key: {},
				},
			},
		},
	})

	node4 := scp.New(scp.Config{
		NodeID:      node4key,
		CurrentSlot: 10,
		Validator:   badApp{},
		Combiner:    badApp{},
		Ledger:      badApp{},
		QuorumSlices: []*scp.QuorumSlice{
			{
				Threshold: 2,
				Validators: map[string]struct{}{
					node2key: {},
					node3key: {},
					node1key: {},
				},
			},
		},
	})

	go node1.Run()
	go node2.Run()
	go node3.Run()
	go node4.Run()

	go func() {
		for {
			m := node1.OutputMessage()
			node2.InputMessage(m)
			node3.InputMessage(m)
			node4.InputMessage(m)
		}
	}()

	go func() {
		for {
			m := node2.OutputMessage()
			node1.InputMessage(m)
			node3.InputMessage(m)
			node4.InputMessage(m)
		}
	}()

	go func() {
		for {
			m := node3.OutputMessage()
			node2.InputMessage(m)
			node1.InputMessage(m)
			node4.InputMessage(m)
		}
	}()

	go func() {
		for {
			m := node4.OutputMessage()
			node2.InputMessage(m)
			node3.InputMessage(m)
			node1.InputMessage(m)
		}
	}()

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
