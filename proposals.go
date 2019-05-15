package scp

type proposals struct {
	input  chan Value
	output chan Value
}

func newProposals() proposals {
	return proposals{
		input: make(chan Value),
	}
}

func (p *proposals) open() {
	p.output = p.input
}

func (p *proposals) close() {
	p.output = nil
}
