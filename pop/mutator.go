package pop

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
