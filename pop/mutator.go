package pop

import (
	"github.com/mingzhi/gsl-cgo/randist"
)

// SimpleMutator implements a simple mutation model,
// which assumes the same mutation rate on each site,
// and equal transition rates between bases.
type SimpleMutator struct {
	// Rand is a source of random numbers
	Rand     Rand
	Alphabet []byte
}

func NewSimpleMutator(r Rand, alphabet []byte) *SimpleMutator {
	return &SimpleMutator{Rand: r, Alphabet: alphabet}
}

func (s *SimpleMutator) Operate(p *Pop) {
	// We need to determine randomly which genome will have a mutation.
	// and which position on the genome.
	g := s.Rand.Intn(p.Size())
	pos := s.Rand.Intn(p.Length())

	// Randomly choose a letter and replace the existed one.
	alphabet := []byte{}
	for j := 0; j < len(s.Alphabet); j++ {
		if s.Alphabet[j] != p.Genomes[g].Seq()[pos] {
			alphabet = append(alphabet, s.Alphabet[j])
		}
	}
	p.Genomes[g].Seq()[pos] = alphabet[s.Rand.Intn(len(alphabet))]
}

type BeneficialMutator struct {
	S    float64
	Rand Rand
}

func (m *BeneficialMutator) Operate(p *Pop) {
	// randomly choose a genome.
	g := m.Rand.Intn(p.Size())
	var ag *NeutralGenome
	// increase its number of beneficial mutation
	ag = p.Genomes[g].(*NeutralGenome)
	ag.fitness += m.S
}

func NewBeneficialMutator(s float64, r Rand) *BeneficialMutator {
	return &BeneficialMutator{Rand: r, S: s}
}

type DeltaMutateFunc func(f *FitnessMutator) (delta float64)

type FitnessMutator struct {
	Scale float64
	Shape float64
	RNG   *randist.RNG
	delta DeltaMutateFunc
}

func NewFitnessMutator(scale, shape float64, rng *randist.RNG, deltaFunc DeltaMutateFunc) *FitnessMutator {
	var f FitnessMutator
	f.Scale = scale
	f.Shape = shape
	f.RNG = rng
	f.delta = deltaFunc
	return &f
}

func (f *FitnessMutator) mutate(p *Pop) {
	g := randist.UniformRandomInt(f.RNG, p.Size())
	var ag *NeutralGenome
	ag = p.Genomes[g].(*NeutralGenome)
	ag.fitness += f.delta(f)
}

func (f *FitnessMutator) Operate(p *Pop) {
	f.mutate(p)
}

func MutateGaussian(f *FitnessMutator) (delta float64) {
	delta = randist.GaussianRandomFloat64(f.RNG, f.Scale)
	return
}

func MutateStep(f *FitnessMutator) (delta float64) {
	delta = f.Scale
	return
}

func MutateExponential(f *FitnessMutator) (delta float64) {
	delta = randist.ExponentialRandomFloat64(f.RNG, f.Scale)
	return delta
}

func MutateGamma(f *FitnessMutator) (delta float64) {
	delta = randist.GammaRandomFloat64(f.RNG, f.Shape, f.Scale)
	return
}
