package pop

import (
	"math/rand"
)

// RouletteWheel is a random generator.
type RouletteWheel struct {
	r *rand.Rand
}

// NewRouletteWheel return a new RouletteWheel.
func NewRouletteWheel(src rand.Source) *RouletteWheel {
	return &RouletteWheel{r: rand.New(src)}
}

// Select return a select index.
func (r *RouletteWheel) Select(weights []float64) (index int) {
	totalWeight := 0.0
	for i := 0; i < len(weights); i++ {
		totalWeight += weights[i]
	}

	v := r.r.Float64()

	accumWeight := 0.0
	for i := 0; i < len(weights); i++ {
		accumWeight += weights[i]
		if accumWeight/totalWeight >= v {
			index = i
			return
		}
	}

	return
}
