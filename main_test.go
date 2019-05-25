package scp_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/xdarksome/scp"
)

type app struct {
	slots map[uint64]scp.Slot
}

func newApp() app {
	return app{
		slots: map[uint64]scp.Slot{
			1: {Index: 1, Value: []byte{1}},
			2: {Index: 2, Value: []byte{2}},
			3: {Index: 3, Value: []byte{3}},
			4: {Index: 4, Value: []byte{4}},
		},
	}
}

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

func (v app) PersistSlot(s scp.Slot) {
	fmt.Println("externalize slot", s.Index, s.Value)
	v.slots[s.Index] = s
}

type catchup struct {
	app
}

func (c catchup) LoadSlots(from uint64, to scp.Slot) (loaded []scp.Slot) {
	time.Sleep(5 * time.Second)
	for i := from; i < to.Index; i++ {
		loaded = append(loaded, c.slots[i])
	}
	return loaded
}

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
		CurrentSlot: 5,
		Validator:   newApp(),
		Combiner:    newApp(),
		Ledger:      newApp(),
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
		CurrentSlot: 5,
		Validator:   newApp(),
		Combiner:    newApp(),
		Ledger:      newApp(),
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

	node3app := newApp()
	node3 := scp.New(scp.Config{
		NodeID:      node3key,
		CurrentSlot: 5,
		Validator:   node3app,
		Combiner:    node3app,
		Ledger:      node3app,
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
		CurrentSlot: 1,
		Validator:   newApp(),
		Combiner:    newApp(),
		Ledger:      newApp(),
		SlotsLoader: &catchup{node3app},
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
			//printMessage(m)
			node2.InputMessage(m)
			node3.InputMessage(m)
			node4.InputMessage(m)
		}
	}()

	go func() {
		for {
			m := node2.OutputMessage()
			//printMessage(m)
			node1.InputMessage(m)
			node3.InputMessage(m)
			node4.InputMessage(m)
		}
	}()

	go func() {
		for {
			m := node3.OutputMessage()
			//printMessage(m)
			node2.InputMessage(m)
			node1.InputMessage(m)
			node4.InputMessage(m)
		}
	}()

	go func() {
		for {
			m := node4.OutputMessage()
			//printMessage(m)
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

func printMessage(m *scp.Message) {
	node := m.NodeID

	var action string
	switch m.Type {
	case scp.VoteNominate:
		action = "votes nominate"
	case scp.AcceptNominate:
		action = "accepts nominate"
	case scp.VotePrepare:
		action = "votes prepare"
	case scp.AcceptPrepare:
		action = "accepts prepare"
	case scp.VoteCommit:
		action = "votes commit"
	case scp.AcceptCommit:
		action = "accepts commit"
	}

	fmt.Println(node, action, "counter:", m.Counter, "value", string(m.Value))
}
