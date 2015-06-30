package pop

import (
	"github.com/mingzhi/gsl-cgo/randist"
	"math"
	"sync"
)

type LinearSelectionSampler struct {
	rng *randist.RNG
	wg  sync.WaitGroup
}

func NewLinearSelectionSampler(rng *randist.RNG) *LinearSelectionSampler {
	var ls LinearSelectionSampler
	ls.rng = rng
	return &ls
}

func (w *LinearSelectionSampler) Operate(p *Pop) {
	defer w.wg.Done()

	meanFit := p.MeanFit()
	sizeRatio := float64(p.Size()) / float64(p.TargetSize)
	// chemical potensial regulating the population size.
	cpot := meanFit - (1.0 - sizeRatio)
	currentGenomes := p.Genomes
	currentLineages := p.Lineages
	newGenomes := []Genome{}
	newLineages := []*Lineage{}
	numGeneration := p.NumGeneration + 1
	for i := 0; i < p.Size(); i++ {
		meanOffSpring := math.Exp(p.Genomes[i].Fitness() - cpot)
		numOffSpring := randist.PoissonRandomInt(w.rng, meanOffSpring)
		for o := 0; o < numOffSpring; o++ {
			var g Genome
			if o == 0 {
				g = currentGenomes[i]
			} else {
				g = currentGenomes[i].Copy()
			}
			newGenomes = append(newGenomes, g)

			l := &Lineage{}
			l.BirthTime = numGeneration
			l.Parent = currentLineages[i]
			newLineages = append(newLineages, l)
		}
	}
	p.Genomes = newGenomes
	p.Lineages = newLineages
	p.NumGeneration = numGeneration
}

func (l *LinearSelectionSampler) Time(p *Pop) float64 {
	return 1.0
}

func (l *LinearSelectionSampler) Start() {
	l.wg.Add(1)
}

func (l *LinearSelectionSampler) Wait() {
	l.wg.Wait()
}
