package quorum

type Slice struct {
	Threshold   uint32
	Validators  []string
	InnerSlices []*InnerSlice
}

type InnerSlice struct {
	Threshold  uint32
	Validators []string
}
