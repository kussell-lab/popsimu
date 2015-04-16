package pop

import (
	"math"
	"math/rand"
	"testing"
)

func TestEmit(t *testing.T) {
	tolerance := 1e-3
	r := rand.New(rand.NewSource(1))
	events := []*Event{
		&Event{Rate: 0.5},
		&Event{Rate: 0.9},
		&Event{Rate: 0},
		&Event{Rate: 0.6},
	}

	eventCountMap := make(map[float64]int)

	numEvents := 1000000
	for i := 0; i < numEvents; i++ {
		e := Emit(events, r)
		eventCountMap[e.Rate]++
	}

	for i := 0; i < len(events); i++ {
		expected := events[i].Rate / 2.0
		result := float64(eventCountMap[events[i].Rate]) / float64(numEvents)
		if math.Abs(result-expected) > tolerance {
			t.Errorf("Expect %f, but obtained %f, within %f\n", expected, result, tolerance)
		}
	}
}
