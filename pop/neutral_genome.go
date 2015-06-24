package pop

type NeutralGenome struct {
	Sequence ByteSequence
	fitness  float64
}

type ByteSequence []byte

func (g *NeutralGenome) Length() int {
	return len(g.Sequence)
}

func (g *NeutralGenome) Seq() []byte {
	return []byte(g.Sequence)
}

func (g *NeutralGenome) Fitness() float64 {
	return g.fitness
}

func (g *NeutralGenome) Copy() Genome {
	var g1 NeutralGenome
	g1.Sequence = make(ByteSequence, g.Length())
	copy(g1.Sequence, g.Sequence)
	g1.fitness = g.fitness
	return &g1
}
