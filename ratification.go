package scp

type ratification struct {
	voted bool
	votes agreement

	accepted bool
	accepts  agreement

	confirmed bool
}

func newRatification() *ratification {
	return &ratification{
		votes:   make(agreement),
		accepts: make(agreement),
	}
}

func (r *ratification) selfVote() {
	r.voted = true
}

func (r *ratification) votedBy(nodeID string) {
	r.votes[nodeID] = struct{}{}
}

func (r *ratification) selfAccept() {
	r.accepted = true
}

func (r *ratification) acceptedBy(nodeID string) {
	r.accepts[nodeID] = struct{}{}
}

func (r *ratification) selfConfirm() {
	r.confirmed = true
}

type agreement map[string]struct{}

func (a agreement) reachedBlockingThreshold(slices quorumSlices) bool {
	for _, slice := range slices {
		if slice.blockingThreshold(a) {
			return true
		}
	}

	return false
}

func (a agreement) reachedQuorumThreshold(slices quorumSlices) bool {
	for _, slice := range slices {
		if slice.quorumThreshold(a) {
			return true
		}
	}

	return false
}

type ballotCounters map[string]uint32

func (c ballotCounters) lowestNonBlocking(current uint32, slices quorumSlices) uint32 {
outer:
	for l := current; ; l++ {
		for _, slice := range slices {
			if slice.blockingCounter(l, c) {
				continue outer
			}
		}
		return l
	}
}

func (c ballotCounters) hasLesserThan(than uint32) bool {
	for _, counter := range c {
		if counter < than {
			return true
		}
	}

	return false
}
