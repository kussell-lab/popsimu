package pop

import (
	"math/rand"

	"github.com/mingzhi/numgo/random"
)

// SimpleMutator implements a simple mutation model,
// which assumes the same mutation rate on each site,
// and equal transition rates between bases.
type SimpleMutator struct {
	// Rand is a source of random numbers
	Alphabet []byte

	r *random.Rand
}

// NewSimpleMutator returns a new SimpleMutator.
func NewSimpleMutator(alphabet []byte, src rand.Source) *SimpleMutator {
	r := random.New(src)
	s := SimpleMutator{Alphabet: alphabet, r: r}
	return &s
}

// Operate mutate a single position at a genome from the *Pop.
func (s *SimpleMutator) Operate(p *Pop) {
	// We need to determine randomly which genome will have a mutation.
	// and which position on the genome.
	g := s.r.Intn(p.Size())
	pos := s.r.Intn(p.Length())

	// Randomly choose a letter and replace the existed one.
	alphabet := []byte{}
	for j := 0; j < len(s.Alphabet); j++ {
		if s.Alphabet[j] != p.Genomes[g].Seq()[pos] {
			alphabet = append(alphabet, s.Alphabet[j])
		}
	}
	p.Genomes[g].Seq()[pos] = alphabet[s.r.Intn(len(alphabet))]
}

// BeneficialMutator is a selective mutator.
type BeneficialMutator struct {
	S float64
	r *random.Rand
}

// Operate increase the fitness score by S.
func (m *BeneficialMutator) Operate(p *Pop) {
	// randomly choose a genome.
	g := m.r.Intn(p.Size())
	var ag *NeutralGenome
	// increase its number of beneficial mutation
	ag = p.Genomes[g].(*NeutralGenome)
	ag.fitness += m.S
}

// NewBeneficialMutator returns a new BeneficialMutator.
func NewBeneficialMutator(s float64, src rand.Source) *BeneficialMutator {
	r := random.New(src)
	return &BeneficialMutator{r: r, S: s}
}

// DeltaMutateFunc is a type of function that increase or decrease
// the fitness score by delta.
type DeltaMutateFunc func(f *FitnessMutator) (delta float64)

// FitnessMutator is a mutator on fitness score.
type FitnessMutator struct {
	Scale float64
	Shape float64
	rand  *random.Rand
	delta DeltaMutateFunc
}

// NewFitnessMutator returns a new fitness mutator.
func NewFitnessMutator(scale, shape float64, src rand.Source, deltaFunc DeltaMutateFunc) *FitnessMutator {
	var f FitnessMutator
	f.Scale = scale
	f.Shape = shape
	f.rand = random.New(src)
	f.delta = deltaFunc
	return &f
}

func (f *FitnessMutator) mutate(p *Pop) {
	g := f.rand.Intn(p.Size())
	var ag *NeutralGenome
	ag = p.Genomes[g].(*NeutralGenome)
	ag.fitness += f.delta(f)
}

// Operate mutate the fitness score.
func (f *FitnessMutator) Operate(p *Pop) {
	f.mutate(p)
}

// FitnessMutateStep return the delta fitness.
func FitnessMutateStep(f *FitnessMutator) (delta float64) {
	delta = f.Scale
	return
}

// FitnessMutateExponential return a exp value of delta.
func FitnessMutateExponential(f *FitnessMutator) (delta float64) {
	delta = f.rand.ExpFloat64(f.Scale)
	return delta
}
