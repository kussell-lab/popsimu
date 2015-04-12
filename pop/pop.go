package pop

import (
	"math/rand"
)

type Pop struct {
	// Size specifies the number of individuals in the population
	Size int
	// Length specifies the length of a genome
	Length int
	// Genomes stores a array of sequences
	Genomes []Sequence
	// Alphabet stores the letters used in constructing the sequences
	Alphabet []byte
	// Circled indicates whether the genome is circled or not.
	Circled bool
}

func New() *Pop {
	return &Pop{}
}

type Sequence []byte

// Objects implementing the Handler interface
// can be used to do operations in the Population.
type Handler interface {
	Operate(*Pop)
}

// RandomPopGenerator randomly generates a population
// with a random ancestor
// given the size of the population
// and the length of the genome.
type RandomPopGenerator struct {
	// Rand is a source of random numbers
	Rand *rand.Rand
}

func NewRandomPopGenerator(r *rand.Rand) *RandomPopGenerator {
	return &RandomPopGenerator{Rand: r}
}

func (r *RandomPopGenerator) Operate(p *Pop) {
	// Randomly generate an acestor.
	ancestor := make(Sequence, p.Length)
	for i := 0; i < p.Length; i++ {
		ancestor[i] = p.Alphabet[r.Rand.Intn(len(p.Alphabet))]
	}

	// Make the genomes and copy the ancestor to each sequence.
	genomes := make([]Sequence, p.Size)
	for i := 0; i < p.Size; i++ {
		genomes[i] = make(Sequence, p.Length)
		copy(genomes[i], ancestor)
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
	Rand *rand.Rand
}

func NewSimpleMutator(rate float64, r *rand.Rand) *SimpleMutator {
	return &SimpleMutator{Rate: rate, Rand: r}
}

func (s *SimpleMutator) Operate(p *Pop) {
	// We need to determine randomly which genome will have a mutation.
	// and which position on the genome.
	g := s.Rand.Intn(p.Size)
	i := s.Rand.Intn(p.Length)
	// Randomly choose a letter and replace the existed one.
	alphabet := []byte{}
	for j := 0; j < len(p.Alphabet); j++ {
		if p.Alphabet[j] != p.Genomes[g][i] {
			alphabet = append(alphabet, p.Alphabet[j])
		}
	}
	p.Genomes[g][i] = alphabet[s.Rand.Intn(len(alphabet))]
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
	Rand *rand.Rand
}

func NewSimpleTransfer(rate float64, fragSize int, r *rand.Rand) *SimpleTransfer {
	return &SimpleTransfer{Rate: rate, FragmentSize: fragSize, Rand: r}
}

func (s *SimpleTransfer) Operate(p *Pop) {
	// We first randomly decise two sequences.
	a := s.Rand.Intn(p.Size)
	b := s.Rand.Intn(p.Size)
	if a != b {
		// Randomly determine the start point of the transfer
		start := s.Rand.Intn(p.Length)
		end := start + s.FragmentSize
		// We need to check whether the end point hits the end of the sequence.
		// And whether is a circled sequence or not.
		if end < p.Length {
			copy(p.Genomes[a][start:end], p.Genomes[b][start:end])
		} else {
			copy(p.Genomes[a][start:p.Length], p.Genomes[b][start:p.Length])
			if p.Circled {
				copy(p.Genomes[a][0:end-p.Length], p.Genomes[b][0:end-p.Length])
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
		copy(p.Genomes[b][start:end], o.DonorPop.Genomes[b][start:end])
	} else {
		copy(p.Genomes[a][start:p.Length], o.DonorPop.Genomes[b][start:p.Length])
		if p.Circled {
			copy(p.Genomes[a][0:end-p.Length], o.DonorPop.Genomes[b][0:end-p.Length])
		}
	}
}

// MoranSampler implements Moran reproduction model.
//
// In each step of Moran process, two individuals are randomly chose:
// one to reproduce and the other to be replaced.
type MoranSampler struct {
	Rand *rand.Rand // random number generator.
}

func NewMoranSampler(r *rand.Rand) *MoranSampler {
	return &MoranSampler{Rand: r}
}

func (m *MoranSampler) Operate(p *Pop) {
	a := m.Rand.Intn(p.Size)
	b := m.Rand.Intn(p.Size)
	if a != b {
		copy(p.Genomes[a], p.Genomes[b])
	}
}
