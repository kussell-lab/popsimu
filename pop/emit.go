package pop

import (
	"sort"
)

type Event struct {
	Ops  Operator
	Rate float64
	Pop  *Pop
}

type ByRate []*Event

func (b ByRate) Len() int           { return len(b) }
func (b ByRate) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByRate) Less(i, j int) bool { return b[i].Rate < b[j].Rate }

// Randomly emit an event accourding to the event rate.
func Emit(events []*Event, r Rand) *Event {
	totalRate := 0.0
	for _, e := range events {
		totalRate += e.Rate
	}

	randomValue := r.Float64()
	sort.Sort(ByRate(events))
	rate := 0.0
	for i := len(events) - 1; i >= 0; i-- {
		rate += events[i].Rate / totalRate
		if randomValue <= rate {
			return events[i]
		}
	}

	return nil
}
