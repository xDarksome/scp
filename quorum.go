package scp

type quorumSlices []*QuorumSlice

type QuorumSlice struct {
	Threshold   uint32
	Validators  map[string]struct{}
	InnerSlices []*QuorumSlice
}

func (q *QuorumSlice) blockingThreshold(a agreement) bool {
	var agreed uint32

	for node := range a {
		if _, ok := q.Validators[node]; ok {
			agreed++
		}
	}

	for _, inner := range q.InnerSlices {
		if inner.blockingThreshold(a) {
			agreed++
		}
	}

	if agreed > uint32(len(q.Validators)+len(q.InnerSlices))-q.Threshold {
		return true
	}

	return false
}

func (q *QuorumSlice) quorumThreshold(a agreement) bool {
	var agreed uint32

	for node := range a {
		if _, ok := q.Validators[node]; ok {
			agreed++
		}
	}

	for _, inner := range q.InnerSlices {
		if inner.quorumThreshold(a) {
			agreed++
		}
	}

	if agreed >= q.Threshold {
		return true
	}

	return false
}

func (q *QuorumSlice) blockingCounter(c uint32, counters map[string]uint32) bool {
	blocking := uint32(0)

	for _, counter := range counters {
		if counter > c {
			blocking++
		}
	}

	for _, inner := range q.InnerSlices {
		if inner.blockingCounter(c, counters) {
			blocking++
		}
	}

	if blocking > uint32(len(q.Validators)+len(q.InnerSlices))-q.Threshold {
		return true
	}

	return false
}

func (q *QuorumSlice) HasNode(id string) bool {
	for v := range q.Validators {
		if v == id {
			return true
		}
	}

	for _, inner := range q.InnerSlices {
		for v := range inner.Validators {
			if v == id {
				return true
			}
		}
	}

	return false
}

func (slices quorumSlices) NodesList() []string {
	set := map[string]struct{}{}
	for _, slice := range slices {
		for v := range slice.Validators {
			set[v] = struct{}{}
		}

		for _, inner := range slice.InnerSlices {
			for v := range inner.Validators {
				set[v] = struct{}{}
			}
		}
	}

	nodes := make([]string, len(set))
	for node := range set {
		nodes = append(nodes, node)
	}

	return nodes
}

func (slices quorumSlices) NodeWeight(id string) (n, d int64) {
	var count int
	for _, slice := range slices {
		if slice.HasNode(id) {
			count++
		}
	}

	return int64(count), int64(len(slices))
}
