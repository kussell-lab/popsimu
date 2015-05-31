package pop

import (
	"sync"
)

type Pop struct {
	// Size specifies the number of individuals in the population
	Size int
	// Length specifies the length of a genome
	Length int
	// Genomes stores a array of sequences
	Genomes []Genome
	// Alphabet stores the letters used in constructing the sequences
	Alphabet []byte
	// Circled indicates whether the genome is circled or not.
	Circled bool

	lk sync.Mutex
}

func New() *Pop {
	return &Pop{}
}

func (p *Pop) LockGenomes(genomeIndex int, optionIndices ...int) {
	p.lk.Lock()
	defer p.lk.Unlock()
	p.Genomes[genomeIndex].Lock()
	for i := 0; i < len(optionIndices); i++ {
		p.Genomes[optionIndices[i]].Lock()
	}
}

func (p *Pop) UnlockGenomes(genomeIndex int, optionIndices ...int) {
	p.Genomes[genomeIndex].Unlock()
	for i := 0; i < len(optionIndices); i++ {
		p.Genomes[optionIndices[i]].Unlock()
	}
}

// Evolve a population by the channel of events.
func Evolve(eventChan chan *Event) {
	for e := range eventChan {
		e.Ops.Operate(e.Pop)
	}
}

// Objects implementing the Operator interface
// can be used to do operations to a population.
type Operator interface {
	Operate(*Pop)
}

type Genome struct {
	lk       sync.Mutex // locker
	Sequence Sequence   // sequence
}

func (g *Genome) Lock() {
	g.lk.Lock()
}

func (g *Genome) Unlock() {
	g.lk.Unlock()
}

type Sequence []byte

// RandomPopGenerator randomly generates a population
// with a random ancestor
// given the size of the population
// and the length of the genome.
type RandomPopGenerator struct {
	// Rand is a source of random numbers
	Rand Rand
}

func NewRandomPopGenerator(r Rand) *RandomPopGenerator {
	return &RandomPopGenerator{Rand: r}
}

func (r *RandomPopGenerator) Operate(p *Pop) {
	// Randomly generate an acestor.
	ancestor := make(Sequence, p.Length)
	for i := 0; i < p.Length; i++ {
		ancestor[i] = p.Alphabet[r.Rand.Intn(len(p.Alphabet))]
	}

	// Make the genomes and copy the ancestor to each sequence.
	genomes := make([]Genome, p.Size)
	for i := 0; i < p.Size; i++ {
		genomes[i].Sequence = make(Sequence, p.Length)
		copy(genomes[i].Sequence, ancestor)
	}

	p.Genomes = genomes
}

// SimplePopGenerator generate a population with the same ancestor.
type SimplePopGenerator struct {
	ancestor Sequence
}

func NewSimplePopGenerator(ancestor Sequence) *SimplePopGenerator {
	return &SimplePopGenerator{ancestor: ancestor}
}

func (s *SimplePopGenerator) Operate(p *Pop) {
	// Make the genomes and copy the ancestor to each sequence.
	genomes := make([]Genome, p.Size)
	for i := 0; i < p.Size; i++ {
		genomes[i].Sequence = make(Sequence, p.Length)
		copy(genomes[i].Sequence, s.ancestor)
	}

	p.Genomes = genomes
}

// SimpleMutator implements a simple mutation model,
// which assumes the same mutation rate on each site,
// and equal transition rates between bases.
type SimpleMutator struct {
	// Rand is a source of random numbers
	Rand Rand
}

func NewSimpleMutator(r Rand) *SimpleMutator {
	return &SimpleMutator{Rand: r}
}

func (s *SimpleMutator) Operate(p *Pop) {
	// We need to determine randomly which genome will have a mutation.
	// and which position on the genome.
	g := s.Rand.Intn(p.Size)
	i := s.Rand.Intn(p.Length)
	// lock the genomes in order to avoid racing.
	p.LockGenomes(g)
	defer p.UnlockGenomes(g)
	// Randomly choose a letter and replace the existed one.
	alphabet := []byte{}
	for j := 0; j < len(p.Alphabet); j++ {
		if p.Alphabet[j] != p.Genomes[g].Sequence[i] {
			alphabet = append(alphabet, p.Alphabet[j])
		}
	}
	p.Genomes[g].Sequence[i] = alphabet[s.Rand.Intn(len(alphabet))]
}

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
	a := m.Rand.Intn(p.Size)
	b := m.Rand.Intn(p.Size)
	if a != b {
		// lock the genomes in order to avoid racing.
		p.LockGenomes(a, b)
		defer p.UnlockGenomes(a, b)
		// copy(p.Genomes[a].Sequence, p.Genomes[b].Sequence)
		for i := 0; i < len(p.Genomes[a].Sequence); i++ {
			if p.Genomes[a].Sequence[i] != p.Genomes[b].Sequence[i] {
				p.Genomes[a].Sequence[i] = p.Genomes[b].Sequence[i]
			}
		}
	}
}
