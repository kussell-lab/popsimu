package pop

import (
	"math/rand"

	"github.com/mingzhi/numgo/random"
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

// NewExpFrag return a new ExpFrag.
func NewExpFrag(lambda float64, src rand.Source) *ExpFrag {
	var e ExpFrag
	e.Lambda = lambda
	e.r = rand.New(src)
	return &e
}

// FragSizeGenerator is a interface with a function return the size of fragment.
type FragSizeGenerator interface {
	Size() int
}

// SimpleTransfer implements a very simple transfer model.
// We randomly choose two sequences:
// one to be the donor, and the other to be the receiver.
// And a piece of the receiver's genome will be replaced by
// a sequence at corresponding genomic positions.
type SimpleTransfer struct {
	Frag FragSizeGenerator

	r *random.Rand
}

func NewSimpleTransfer(frag FragSizeGenerator, src rand.Source) *SimpleTransfer {
	return &SimpleTransfer{Frag: frag, r: random.New(src)}
}

func (s *SimpleTransfer) Operate(p *Pop) {
	var length int
	// We first randomly decise two sequences.
	a := s.r.Intn(p.Size())
	b := s.r.Intn(p.Size())
	length = p.Genomes[a].Length()
	if a != b {
		// Randomly determine the start point of the transfer
		start := s.r.Intn(length)
		end := start + s.Frag.Size()
		// We need to check whether the end point hits the end of the sequence.
		// And whether is a circled sequence or not.
		if end < length {
			copy(p.Genomes[a].Seq()[start:end], p.Genomes[b].Seq()[start:end])
		} else {
			copy(p.Genomes[a].Seq()[start:length], p.Genomes[b].Seq()[start:length])
			if p.Circled {
				copy(p.Genomes[a].Seq()[0:end-length], p.Genomes[b].Seq()[0:end-length])
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

// NewOutTransfer return a new OutTransfer.
func NewOutTransfer(frag FragSizeGenerator, donorPop *Pop, src rand.Source) *OutTransfer {
	o := OutTransfer{}
	o.r = random.New(src)
	o.Frag = frag
	o.DonorPop = donorPop
	return &o
}

// Operate perform a out population transfer.
func (o *OutTransfer) Operate(p *Pop) {
	var length int // genome length
	// We first randomly choose a sequence from the donor sequence,
	// and a sequence from the receipient population.
	a := o.r.Intn(o.DonorPop.Size())
	b := o.r.Intn(p.Size())

	length = p.Genomes[b].Length()

	// Randomly determine the start point of the transfer.
	start := o.r.Intn(length)
	end := start + o.Frag.Size()
	// We need to check whether the point hits the boundary of the sequence.
	if end < length {
		copy(p.Genomes[b].Seq()[start:end], o.DonorPop.Genomes[a].Seq()[start:end])
	} else {
		copy(p.Genomes[b].Seq()[start:length], o.DonorPop.Genomes[a].Seq()[start:length])
		if p.Circled {
			copy(p.Genomes[b].Seq()[0:end-length], o.DonorPop.Genomes[a].Seq()[0:end-length])
		}
	}
}
