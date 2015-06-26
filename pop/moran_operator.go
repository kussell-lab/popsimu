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

func initLineaages(p *Pop) {
	p.Lineages = make([]*Lineage, p.Size())
	for i := 0; i < p.Size(); i++ {
		p.Lineages[i] = &Lineage{}
	}
}

func createNewLineages(parent *Lineage, t int) (a, b *Lineage) {
	a = &Lineage{}
	b = &Lineage{}
	a.Parent = parent
	b.Parent = parent
	a.BirthTime = t
	b.BirthTime = t

	return
}

func (m *MoranSampler) Operate(p *Pop) {
	if len(p.Lineages) < p.Size() {
		initLineaages(p)
	}
	p.NumGeneration++

	// random choose a going-death one
	d := m.Rand.Intn(p.Size())
	// random choose a going-birth one according to the fitness.
	var weights []float64
	for i := 0; i < p.Size(); i++ {
		var f float64
		f = p.Genomes[i].Fitness()
		weights = append(weights, f+1.0/float64(p.Size()))
	}
	b := RouletteWheelSelect(weights, m.Rand)

	if d != b {
		p.Genomes[d] = p.Genomes[b].Copy()
	}

	p.Lineages[b], p.Lineages[d] = createNewLineages(p.Lineages[b], p.NumGeneration)
}
