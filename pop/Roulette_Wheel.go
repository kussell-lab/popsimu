package pop

func RouletteWheelSelect(weights []float64, r Rand) (index int) {
	totalWeight := 0.0
	for i := 0; i < len(weights); i++ {
		totalWeight += weights[i]
	}

	v := r.Float64()

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
