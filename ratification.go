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

/*

type nominateRatification struct {
	Value

	Voted bool

	Votes    thresholds
	Accepted bool

	Accepts   thresholds
	Confirmed bool
}

func newNominateRatification(v Value, slices quorumSlices) *nominateRatification {
	return &nominateRatification{
		Value:   v,
		Votes:   newThresholds(slices),
		Accepts: newThresholds(slices),
	}
}

func (n *nominateRatification) setThresholds(slices quorumSlices) {
	n.Votes = newThresholds(slices)
	n.Accepts = newThresholds(slices)
}

func (n *nominateRatification) Less(than btree.Item) bool {
	return bytes.Compare(n.Value, than.(*nominateRatification).Value) == -1
}

func (n *nominateRatification) Vote() {
	n.Votes.self()
	n.Voted = true
}

func (n *nominateRatification) Accept() {
	n.Accepts.self()
	n.Accepted = true
}

func (n *nominateRatification) Confirm() {
	n.Confirmed = true
}

func (n *nominateRatification) VotedBy(id string) {
	n.Votes.mark(id)
}

func (n *nominateRatification) AcceptedBy(id string) {
	n.VotedBy(id)
	n.Accepts.mark(id)
}

type ballotRatification struct {
	ballot

	PrepareVoted bool

	PrepareVotes    thresholds
	PrepareAccepted bool

	PrepareAccepts   thresholds
	PrepareConfirmed bool

	CommitVoted bool

	CommitVotes    thresholds
	CommitAccepted bool

	CommitAccepts   thresholds
	CommitConfirmed bool
}

func newBallotRatification(b ballot) *ballotRatification {
	return &ballotRatification{
		ballot: b,
	}
}

func (b *ballotRatification) setThresholds(slices quorumSlices) *ballotRatification {
	b.PrepareVotes = newThresholds(slices)
	b.PrepareAccepts = newThresholds(slices)
	b.CommitVotes = newThresholds(slices)
	b.CommitAccepts = newThresholds(slices)
	return b
}

func (b *ballotRatification) VotePrepare() {
	b.PrepareVotes.self()
	b.PrepareVoted = true
}

func (b *ballotRatification) PrepareVotedBy(nodeID string) {
	b.PrepareVotes.mark(nodeID)
}

func (b *ballotRatification) AcceptPrepare() {
	b.PrepareAccepts.self()
	b.PrepareAccepted = true
}

func (b *ballotRatification) PrepareAcceptedBy(nodeID string) {
	b.PrepareAccepts.mark(nodeID)
}

func (b *ballotRatification) ConfirmPrepare() {
	b.PrepareConfirmed = true
}

func (b *ballotRatification) CommitVotedBy(nodeID string) {
	b.CommitVotes.mark(nodeID)
}

func (b *ballotRatification) CommitAcceptedBy(nodeID string) {
	b.CommitAccepts.mark(nodeID)
}

type thresholds []*threshold

func newThresholds(slices quorumSlices) (t thresholds) {
	for _, slice := range slices {
		t = append(t, newThreshold(slice))
	}
	return t
}

func (t thresholds) ReachedQuorumThreshold() bool {
	for _, threshold := range []*threshold(t) {
		if threshold.reachedQuorum() {
			return true
		}
	}

	return false
}

func (t thresholds) ReachedBlockingThreshold() bool {
	for _, threshold := range []*threshold(t) {
		if threshold.reachedBlocking() {
			return true
		}
	}

	return false
}

func (t thresholds) self() {
	for _, threshold := range []*threshold(t) {
		threshold.markSelf()
	}
}

func (t thresholds) mark(nodeID string) {
	for _, threshold := range []*threshold(t) {
		threshold.markNode(nodeID)
	}
}

type threshold struct {
	total    uint32
	expected uint32
	current  uint32
	self     bool

	pending      map[string]struct{}
	pendingInner []*innerThreshold
}

func (t *threshold) reachedQuorum() bool {
	return t.self && t.current >= t.expected
}

func (t *threshold) reachedBlocking() bool {
	return t.current >= t.total-t.expected
}

func (t *threshold) markSelf() {
	if t.self {
		return
	}

	t.self = true
	t.current++
}

func (t *threshold) markNode(id string) {
	if _, ok := t.pending[id]; ok {
		t.current++
		delete(t.pending, id)
	}

	for i, inner := range t.pendingInner {
		inner.nodeAgreed(id)
		if !inner.reachedQuorumThreshold() {
			continue
		}

		t.current++
		t.pendingInner[i] = t.pendingInner[len(t.pendingInner)-1]
		t.pendingInner[len(t.pendingInner)-1] = nil
		t.pendingInner = t.pendingInner[:len(t.pendingInner)-1]
	}
}

type innerThreshold struct {
	total    uint32
	expected uint32
	current  uint32
	pending  map[string]struct{}
}

func (i *innerThreshold) nodeAgreed(id string) {
	if _, ok := i.pending[id]; ok {
		i.current++
		delete(i.pending, id)
	}
}

func (i *innerThreshold) reachedQuorumThreshold() bool {
	return i.current >= i.expected
}

func (i *innerThreshold) reachedBlockingThreshold() bool {
	return i.current >= i.total-i.expected
}

func newThreshold(s *QuorumSlice) *threshold {
	t := threshold{}
	t.expected = s.Threshold
	t.pending = map[string]struct{}{}
	for _, id := range s.Validators {
		t.pending[id.String()] = struct{}{}
	}

	for _, inner := range s.InnerSlices {
		it := innerThreshold{
			pending: map[string]struct{}{},
		}
		for _, id := range inner.Validators {
			it.pending[id.String()] = struct{}{}
		}
		t.pendingInner = append(t.pendingInner, &it)
	}

	t.total = uint32(len(t.pending) + len(t.pendingInner))

	return &t
}

type ballotCounterThresholds []*ballotCounterThreshold

func newBallotCounterThresholds(slices quorumSlices) ballotCounterThresholds {
	t := make([]*ballotCounterThreshold, len(slices))

	for _, slice := range slices {
		t = append(t, newQuorumSliceBallotCounters(slice))
	}

	return ballotCounterThresholds(t)
}

func (t ballotCounterThresholds) SetCounter(nodeID string, counter uint32) {
	for _, threshold := range t {
		threshold.setCounter(nodeID, counter)
	}
}

func (t ballotCounterThresholds) HasCounterLessThan(c uint32) bool {
	for _, threshold := range t {
		if threshold.hasCounterLessThan(c) {
			return true
		}
	}

	return false
}

func (t ballotCounterThresholds) LowestNonBlockingCounter(local uint32) uint32 {
outer:
	for l := local; ; l++ {
		for _, threshold := range t {
			if threshold.isBlocking(l) {
				continue outer
			}
		}
		return l
	}
}

type ballotCounterThreshold struct {
	threshold uint32
	total     uint32
	counters  map[string]uint32
	inner     []*ballotCounterThreshold
}

func newQuorumSliceBallotCounters(s *QuorumSlice) *ballotCounterThreshold {
	c := &ballotCounterThreshold{
		threshold: s.Threshold,
		total:     uint32(len(s.Validators) + len(s.InnerSlices)),
		counters:  map[string]uint32{},
	}
	for _, v := range s.Validators {
		c.counters[v.String()] = 0
	}

	for _, inner := range s.InnerSlices {
		i := &ballotCounterThreshold{
			threshold: inner.Threshold,
			total:     uint32(len(inner.Validators)),
			counters:  map[string]uint32{},
		}
		for _, v := range inner.Validators {
			i.counters[v.String()] = 0
		}
		c.inner = append(c.inner, i)
	}

	return c
}

func (q *ballotCounterThreshold) setCounter(nodeID string, counter uint32) {
	q.counters[nodeID] = counter
	for _, inner := range q.inner {
		inner.setCounter(nodeID, counter)
	}
}

func (q *ballotCounterThreshold) hasCounterLessThan(c uint32) bool {
	for _, counter := range q.counters {
		if counter < c {
			return true
		}
	}

	for _, inner := range q.inner {
		if inner.hasCounterLessThan(c) {
			return true
		}
	}

	return false
}

func (q *ballotCounterThreshold) isBlocking(c uint32) bool {
	blocking := uint32(0)

	for _, counter := range q.counters {
		if counter > c {
			blocking++
		}
	}

	for _, inner := range q.inner {
		if inner.isBlocking(c) {
			blocking++
		}
	}

	if blocking > q.total-q.threshold {
		return true
	}

	return false
}
*/
