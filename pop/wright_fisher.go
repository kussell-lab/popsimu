package pop

import (
	"github.com/mingzhi/gsl-cgo/randist"
	"sync"
)

// WrightFisherSampler for Wright-Fisher reproduction model.
type WrightFisherSampler struct {
	rng *randist.RNG
	wg  sync.WaitGroup
}

// NewWrightFisherSampler create a new WrightFisherSampler.
func NewWrightFisherSampler(r *randist.RNG) *WrightFisherSampler {
	var w WrightFisherSampler
	w.rng = r
	return &w
}

func (w *WrightFisherSampler) Operate(p *Pop) {
	defer w.wg.Done()

	if len(p.Lineages) < p.Size() {
		p.NewLineages()
	}
	currentGenomes := p.Genomes
	currentLineages := p.Lineages
	newGenomes := make([]Genome, p.Size())
	newLineages := make([]*Lineage, p.Size())
	newGeneration := p.NumGeneration + 1

	usedGenomes := make(map[int]bool)
	for i := 0; i < p.Size(); i++ {
		index := randist.UniformRandomInt(w.rng, p.Size())
		if usedGenomes[index] {
			newGenomes[i] = currentGenomes[index].Copy()
		} else {
			newGenomes[i] = currentGenomes[index]
		}
		usedGenomes[index] = true

		newLineages[i] = &Lineage{}
		newLineages[i].BirthTime = newGeneration
		newLineages[i].Parent = currentLineages[index]
	}

	p.Genomes = newGenomes
	p.Lineages = newLineages
	p.NumGeneration = newGeneration
}

func (w *WrightFisherSampler) Time(p *Pop) float64 {
	return 1.0
}

func (w *WrightFisherSampler) Start() {
	w.wg.Add(1)
}

func (w *WrightFisherSampler) Wait() {
	w.wg.Wait()
}
