package pop

// MoranSampler implements Moran reproduction model.
//
// In each step of Moran process, two individuals are randomly chose:
// one to reproduce and the other to be replaced.
type MoranSampler struct {
	Rand Rand // random number generator.
}

func NewMoranSampler(r Rand) *MoranSampler {
	return &MoranSampler{Rand: r}
}

func (m *MoranSampler) Operate(p *Pop) {
	// random choose a going-death one
	d := m.Rand.Intn(p.Size())
	// random choose a going-birth one according to the fitness.
	var weights []float64
	for i := 0; i < p.Size(); i++ {
		var f float64
		f = p.Genomes[i].Fitness()
		weights = append(weights, f)
	}
	b := RouletteWheelSelect(weights, m.Rand)

	if d != b {
		p.Genomes[d] = p.Genomes[b].Copy()
	}
}
