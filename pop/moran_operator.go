package pop

import (
	"math"
	"sync"
	"math/rand"
	"github.com/mingzhi/numgo/random"
)

// MoranSampler implements Moran reproduction model.
//
// In each step of Moran process, two individuals are randomly chose:
// one to reproduce and the other to be replaced.
type MoranSampler struct {
	rng *random.Rand// random number generator.
	wg  sync.WaitGroup
}

func NewMoranSampler(src rand.Source) *MoranSampler {
	return &MoranSampler{rng: random.New(src)}
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
	defer m.wg.Done()
	if len(p.Lineages) < p.Size() {
		p.NewLineages()
	}
	p.NumGeneration++

	// random choose a going-death one
	d := rand.Intn(p.Size())
	// random choose a going-birth one according to the fitness.
	meanFit := p.MeanFit()
	var weights []float64
	for i := 0; i < p.Size(); i++ {
		meanOffSpring := math.Exp(p.Genomes[i].Fitness() - meanFit)
		// var f float64
		// f = p.Genomes[i].Fitness()
		// weights = append(weights, f+1.0/float64(p.Size()))
		weights = append(weights, meanOffSpring)
	}
	b := RouletteWheelSelect(weights)

	if d != b {
		p.Genomes[d] = p.Genomes[b].Copy()
	}

	p.Lineages[b], p.Lineages[d] = createNewLineages(p.Lineages[b], p.NumGeneration)
}

func (m *MoranSampler) Time(p *Pop) float64 {
	lambda := 1 / float64(p.Size())
	t := m.rng.Exponential(lambda)
	return t
}

func (m *MoranSampler) Wait() {
	m.wg.Wait()
}

func (m *MoranSampler) Start() {
	m.wg.Add(1)
}
