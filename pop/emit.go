package pop

type Event struct {
	Ops  Operator
	Rate float64
	Pop  *Pop
}

type ByRate []*Event

func (b ByRate) Len() int           { return len(b) }
func (b ByRate) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByRate) Less(i, j int) bool { return b[i].Rate > b[j].Rate }

// Emit randomly emits an event accourding to the event rate.
func Emit(events []*Event, rw *RouletteWheel) *Event {
	var weights []float64
	for _, e := range events {
		weights = append(weights, e.Rate)
	}
	index := rw.Select(weights)
	return events[index]
}
