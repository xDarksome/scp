package scp

type candidates struct {
	queue   chan Value
	current Value
}

func newCandidates() candidates {
	return candidates{
		queue: make(chan Value, 1000),
	}
}
