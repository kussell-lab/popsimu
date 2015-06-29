package pop

import (
	"math"
)

type Pop struct {
	// Genomes stores a array of sequences
	Genomes []Genome
	// Circled indicates whether the genome is circled or not.
	Circled       bool
	Lineages      []*Lineage
	NumGeneration int
	TargetSize    int
}

func New() *Pop {
	return &Pop{}
}

func (p *Pop) Size() int {
	return len(p.Genomes)
}

func (p *Pop) Length() int {
	if len(p.Genomes) == 0 {
		return 0
	} else {
		return p.Genomes[0].Length()
	}
}

func (p *Pop) NewLineages() {
	p.Lineages = make([]*Lineage, p.Size())
	for i := 0; i < p.Size(); i++ {
		p.Lineages[i] = &Lineage{}
	}
}

func (p *Pop) MeanFit() float64 {
	var m float64
	for i := 0; i < p.Size(); i++ {
		m += p.Genomes[i].Fitness()
	}
	return m / float64(p.Size())
}

func (p *Pop) MaxFit() float64 {
	var max float64 = math.Inf(-1)
	for i := 0; i < p.Size(); i++ {
		if max < p.Genomes[i].Fitness() {
			max = p.Genomes[i].Fitness()
		}
	}
	return max
}

// RandomPopGenerator randomly generates a population
// with a random neutral ancestral genome,
// given the size of the population
// and the length of the genome.
type RandomPopGenerator struct {
	// Rand is a source of random numbers
	Rand     Rand
	Alphabet []byte
	Size     int // size of population
	Length   int // length of genome
}

func NewRandomPopGenerator(r Rand,
	size, length int,
	alphabet []byte) *RandomPopGenerator {

	return &RandomPopGenerator{
		Rand:     r,
		Size:     size,
		Length:   length,
		Alphabet: alphabet,
	}
}

func (r *RandomPopGenerator) Operate(p *Pop) {
	// Randomly generate an acestral sequence.
	ancestor := make(ByteSequence, r.Length)
	for i := 0; i < r.Length; i++ {
		ancestor[i] = r.Alphabet[r.Rand.Intn(len(r.Alphabet))]
	}

	// Make the genomes and copy the ancestor to each sequence.
	genomes := make([]NeutralGenome, r.Size)
	for i := 0; i < r.Size; i++ {
		genomes[i].Sequence = make(ByteSequence, r.Length)
		copy(genomes[i].Sequence, ancestor)
	}

	p.Genomes = make([]Genome, r.Size)
	for i := 0; i < len(genomes); i++ {
		p.Genomes[i] = &genomes[i]
	}

	p.TargetSize = r.Size
	p.NewLineages()
}

// SimplePopGenerator generate a population with the same ancestral sequence.
type SimplePopGenerator struct {
	Ancestor Genome
	Size     int // size of population
}

func NewSimplePopGenerator(ancestor Genome, size int) *SimplePopGenerator {
	return &SimplePopGenerator{Ancestor: ancestor, Size: size}
}

func (s *SimplePopGenerator) Operate(p *Pop) {
	for i := 0; i < s.Size; i++ {
		var g Genome
		g = s.Ancestor.Copy()
		p.Genomes = append(p.Genomes, g)
	}
}
