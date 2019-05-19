package scp

type catchuper struct {
	loader SlotsLoader
	ledger Ledger
	from   uint64
	to     Slot
	done   chan struct{}
}

func (c catchuper) catchUp() {
	slots := c.loader.LoadSlots(c.from, c.to)
	for _, s := range slots {
		c.ledger.PersistSlot(s)
	}
	c.done <- struct{}{}
}
