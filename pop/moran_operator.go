package pop

type Lineage struct {
	Id     int // index at the current generation.
	Time   int // when it was produced.
	Parent *Lineage
}

// MoranSampler implements Moran reproduction model.
//
// In each step of Moran process, two individuals are randomly chose:
// one to reproduce and the other to be replaced.
type MoranSampler struct {
	Rand     Rand // random number generator.
	Lineages []*Lineage
	Time     int
}

func NewMoranSampler(r Rand) *MoranSampler {
	return &MoranSampler{Rand: r}
}

func (m *MoranSampler) Operate(p *Pop) {
	m.Time++
	if len(m.Lineages) < p.Size() {
		m.Lineages = make([]*Lineage, p.Size())
		for i := 0; i < p.Size(); i++ {
			var l Lineage
			l.Id = i
			l.Time = 0
			l.Parent = nil
			m.Lineages[i] = &l
		}
	}

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

	var parent *Lineage
	parent = m.Lineages[b]

	var al, bl Lineage
	al.Id = b
	al.Time = m.Time
	al.Parent = parent
	bl.Id = d
	bl.Time = m.Time
	bl.Parent = parent
	m.Lineages[b] = &al
	m.Lineages[d] = &bl
}
