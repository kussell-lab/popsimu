package pop

// Evolve a population by the channel of events.
func Evolve(eventChan chan *Event) {
	for e := range eventChan {
		e.Ops.Operate(e.Pop)
	}
}
