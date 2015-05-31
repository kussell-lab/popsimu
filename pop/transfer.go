package pop

import (
	"math/rand"
)

type ConstantFrag struct {
	Length int
}

func (c *ConstantFrag) Size() int {
	return c.Length
}

func NewConstantFrag(length int) *ConstantFrag {
	var c ConstantFrag
	c.Length = length
	return &c
}

type ExpFrag struct {
	Lambda float64
	r      *rand.Rand
}

func (e *ExpFrag) Size() int {
	return int(e.r.ExpFloat64() / e.Lambda)
}

func NewExpFrag(lambda float64, src rand.Source) *ExpFrag {
	var e ExpFrag
	e.Lambda = lambda
	e.r = rand.New(src)
	return &e
}

type FragSizeGenerator interface {
	Size() int
}

// SimpleTransfer implements a very simple transfer model.
// We randomly choose two sequences:
// one to be the donor, and the other to be the receiver.
// And a piece of the receiver's genome will be replaced by
// a sequence at corresponding genomic positions.
type SimpleTransfer struct {
	// Rand is a source of random numbers.
	Rand Rand
	Frag FragSizeGenerator
}

func NewSimpleTransfer(frag FragSizeGenerator, r Rand) *SimpleTransfer {
	return &SimpleTransfer{Frag: frag, Rand: r}
}

func (s *SimpleTransfer) Operate(p *Pop) {
	// We first randomly decise two sequences.
	a := s.Rand.Intn(p.Size)
	b := s.Rand.Intn(p.Size)
	if a != b {
		// lock the genomes in order to avoid racing.
		// p.LockGenomes(a, b)
		// defer p.UnlockGenomes(a, b)
		// Randomly determine the start point of the transfer
		start := s.Rand.Intn(p.Length)
		end := start + s.Frag.Size()
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

func NewOutTransfer(frag FragSizeGenerator, donorPop *Pop, r Rand) *OutTransfer {
	o := OutTransfer{}
	o.Rand = r
	o.Frag = frag
	o.DonorPop = donorPop
	return &o
}

func (o *OutTransfer) Operate(p *Pop) {
	// We first randomly choose a sequence from the donor sequence,
	// and a sequence from the receipient population.
	a := o.Rand.Intn(o.DonorPop.Size)
	b := o.Rand.Intn(p.Size)

	// Randomly determine the start point of the transfer.
	start := o.Rand.Intn(p.Length)
	end := start + o.Frag.Size()
	// We need to check whether the point hits the boundary of the sequence.
	if end < p.Length {
		copy(p.Genomes[b].Sequence[start:end], o.DonorPop.Genomes[a].Sequence[start:end])
	} else {
		copy(p.Genomes[b].Sequence[start:p.Length], o.DonorPop.Genomes[a].Sequence[start:p.Length])
		if p.Circled {
			copy(p.Genomes[b].Sequence[0:end-p.Length], o.DonorPop.Genomes[a].Sequence[0:end-p.Length])
		}
	}
}
