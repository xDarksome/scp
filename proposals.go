package scp

type proposals struct {
	open   bool
	input  chan Value
	close  chan struct{}
	output chan Value
}

func newProposals() proposals {
	return proposals{
		input:  make(chan Value),
		close:  make(chan struct{}),
		output: make(chan Value),
	}
}

func (p *proposals) start() {
	p.open = true
	go func() {
		for {
			select {
			case v := <-p.input:
				p.output <- v
			case <-p.close:
				p.open = false
				return
			}
		}
	}()
}

func (p *proposals) stop() {
	if p.open {
		p.close <- struct{}{}
	}
}
