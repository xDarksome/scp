package scp

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

type protocolMessage struct {
	slotIndex  uint64
	candidates chan Value
}
