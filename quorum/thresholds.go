package quorum

type thresholds []*Threshold

func newThresholds(s *Slices) (t thresholds) {
	for _, slice := range s.List {
		t = append(t, newThreshold(slice))
	}
	return t
}

func (t thresholds) ReachedQuorumThreshold() bool {
	for _, threshold := range []*Threshold(t) {
		if threshold.reachedQuorum() {
			return true
		}
	}

	return false
}

func (t thresholds) ReachedBlockingThreshold() bool {
	for _, threshold := range []*Threshold(t) {
		if threshold.reachedBlocking() {
			return true
		}
	}

	return false
}

func (t thresholds) self() {
	for _, threshold := range []*Threshold(t) {
		threshold.markSelf()
	}
}

func (t thresholds) mark(nodeID string) {
	for _, threshold := range []*Threshold(t) {
		threshold.markNode(nodeID)
	}
}

type Threshold struct {
	total    uint32
	expected uint32
	current  uint32
	self     bool

	pending      map[string]struct{}
	pendingInner []*innerThreshold
}

func (t *Threshold) reachedQuorum() bool {
	return t.self && t.current >= t.expected
}

func (t *Threshold) reachedBlocking() bool {
	return t.current >= t.total-t.expected
}

func (t *Threshold) markSelf() {
	if t.self {
		return
	}

	t.self = true
	t.current++
}

func (t *Threshold) markNode(id string) {
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

func newThreshold(s *Slice) *Threshold {
	t := Threshold{}
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

type BallotCounters struct {
	slices []*sliceBallotCounters
}

func newQuorumBallotCounters(s *Slices) *BallotCounters {
	slices := make([]*sliceBallotCounters, len(s.List))

	for _, slice := range s.List {
		slices = append(slices, newQuorumSliceBallotCounters(slice))
	}

	return &BallotCounters{
		slices: slices,
	}
}

func (q *BallotCounters) set(nodeID string, counter uint32) {
	for _, slice := range q.slices {
		slice.set(nodeID, counter)
	}
}

func (q *BallotCounters) hasLessThan(c uint32) bool {
	for _, slice := range q.slices {
		if slice.hasLessThan(c) {
			return true
		}
	}

	return false
}

func (q *BallotCounters) lowestNonBlocking(local uint32) uint32 {
outer:
	for l := local; ; l++ {
		for _, slice := range q.slices {
			if slice.isBlocking(l) {
				continue outer
			}
		}
		return l
	}
}

type sliceBallotCounters struct {
	threshold uint32
	total     uint32
	counters  map[string]uint32
	inner     []*sliceBallotCounters
}

func newQuorumSliceBallotCounters(s *Slice) *sliceBallotCounters {
	c := &sliceBallotCounters{
		threshold: s.Threshold,
		total:     uint32(len(s.Validators) + len(s.InnerSlices)),
		counters:  map[string]uint32{},
	}
	for _, v := range s.Validators {
		c.counters[v.String()] = 0
	}

	for _, inner := range s.InnerSlices {
		i := &sliceBallotCounters{
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

func (q *sliceBallotCounters) set(nodeID string, counter uint32) {
	q.counters[nodeID] = counter
	for _, inner := range q.inner {
		inner.set(nodeID, counter)
	}
}

func (q *sliceBallotCounters) hasLessThan(c uint32) bool {
	for _, counter := range q.counters {
		if counter < c {
			return true
		}
	}

	for _, inner := range q.inner {
		if inner.hasLessThan(c) {
			return true
		}
	}

	return false
}

func (q *sliceBallotCounters) isBlocking(c uint32) bool {
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
