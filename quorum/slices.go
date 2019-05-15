package quorum

import (
	"github.com/xdarksome/scp/key"
)

type Slices struct {
	self key.Public
	List []*Slice
}

func (s *Slices) Self() string {
	return s.self.String()
}

func (s *Slices) FindAlias(id string) string {
	if id == s.self.String() {
		return s.self.Alias
	}

	for _, slice := range s.List {
		for _, v := range slice.Validators {
			if v.String() == id {
				return v.Alias
			}
		}

		for _, inner := range slice.InnerSlices {
			for _, v := range inner.Validators {
				if v.String() == id {
					return v.Alias
				}
			}
		}

	}

	return ""
}

type Slice struct {
	Threshold   uint32
	Validators  map[string]key.Public
	InnerSlices []*InnerSlice
}

type InnerSlice struct {
	Threshold  uint32
	Validators map[string]key.Public
}

func (q *Slice) HasNode(id string) bool {
	for _, v := range q.Validators {
		if v.String() == id {
			return true
		}
	}

	for _, inner := range q.InnerSlices {
		for _, v := range inner.Validators {
			if v.String() == id {
				return true
			}
		}
	}

	return false
}

func (s Slices) NodesList() []string {
	set := map[string]struct{}{}
	for _, slice := range s.List {
		for _, validator := range slice.Validators {
			set[validator.String()] = struct{}{}
		}

		for _, inner := range slice.InnerSlices {
			for _, validator := range inner.Validators {
				set[validator.String()] = struct{}{}
			}
		}
	}

	nodes := make([]string, len(set))
	for node := range set {
		nodes = append(nodes, node)
	}

	return nodes
}

func (s Slices) NodeWeight(id string) (n, d int64) {
	var count int
	for _, slice := range s.List {
		if slice.HasNode(id) {
			count++
		}
	}

	return int64(count), int64(len(s.List))
}
