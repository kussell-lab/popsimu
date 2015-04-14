package pop

import (
	"runtime"
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

// Evolve a population following the operations.
func Evolve(p *Pop, operations chan Operator) {
	ncpu := runtime.GOMAXPROCS(0)
	done := make(chan bool)
	for i := 0; i < ncpu; i++ {
		go func() {
			for ops := range operations {
				ops.Operate(p)
			}
			done <- true
		}()
	}

	for i := 0; i < ncpu; i++ {
		<-done
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
	// Rate sepcifies the mutation rate
	Rate float64
	// Rand is a source of random numbers
	Rand Rand
}

func NewSimpleMutator(rate float64, r Rand) *SimpleMutator {
	return &SimpleMutator{Rate: rate, Rand: r}
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

// SimpleTransfer implements a very simple transfer model.
// We randomly choose two sequences:
// one to be the donor, and the other to be the receiver.
// And a piece of the receiver's genome will be replaced by
// a sequence at corresponding genomic positions.
type SimpleTransfer struct {
	// Rate specifies the transfer rate.
	Rate float64
	// FragmentSize denotes the length of transferred fragments.
	FragmentSize int
	// Rand is a source of random numbers.
	Rand Rand
}

func NewSimpleTransfer(rate float64, fragSize int, r Rand) *SimpleTransfer {
	return &SimpleTransfer{Rate: rate, FragmentSize: fragSize, Rand: r}
}

func (s *SimpleTransfer) Operate(p *Pop) {
	// We first randomly decise two sequences.
	a := s.Rand.Intn(p.Size)
	b := s.Rand.Intn(p.Size)
	if a != b {
		// lock the genomes in order to avoid racing.
		p.LockGenomes(a, b)
		defer p.UnlockGenomes(a, b)
		// Randomly determine the start point of the transfer
		start := s.Rand.Intn(p.Length)
		end := start + s.FragmentSize
		// We need to check whether the end point hits the end of the sequence.
		// And whether is a circled sequence or not.
		if end < p.Length {
			copy(p.Genomes[a].Sequence[start:end], p.Genomes[b].Sequence[start:end])
		} else {
			copy(p.Genomes[a].Sequence[start:p.Length], p.Genomes[b].Sequence[start:p.Length])
			if p.Circled {
				copy(p.Genomes[a].Sequence[0:end-p.Length], p.Genomes[b].Sequence[0:end-p.Length])
			}
		}
	}
}

// OutTransfer implements transfers from a donor population to a receiver one.
//
// We randomly choose a sequence from the donor population,
// and one from the reciever population.
type OutTransfer struct {
	SimpleTransfer
	DonorPop *Pop
}

func (o *OutTransfer) Operate(p *Pop) {
	// We first randomly choose a sequence from the donor sequence,
	// and a sequence from the receipient population.
	a := o.Rand.Intn(o.DonorPop.Size)
	b := o.Rand.Intn(p.Size)

	// Randomly determine the start point of the transfer.
	start := o.Rand.Intn(p.Length)
	end := start + o.FragmentSize
	// We need to check whether the point hits the boundary of the sequence.
	if end < p.Length {
		copy(p.Genomes[b].Sequence[start:end], o.DonorPop.Genomes[b].Sequence[start:end])
	} else {
		copy(p.Genomes[a].Sequence[start:p.Length], o.DonorPop.Genomes[b].Sequence[start:p.Length])
		if p.Circled {
			copy(p.Genomes[a].Sequence[0:end-p.Length], o.DonorPop.Genomes[b].Sequence[0:end-p.Length])
		}
	}
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
		copy(p.Genomes[a].Sequence, p.Genomes[b].Sequence)
	}
}
