package pop

import (
	"bytes"
	"fmt"
)

// Population config
type Config struct {
	// population parameters
	Size     int // population size
	Length   int // length of genome
	NumGen   int // number of generations.
	Alphabet string

	Mutation struct {
		Rate float64
	}

	Transfer struct {
		In struct {
			Rate     float64
			Fragment int
		}
		Out struct {
			Rate     float64
			Fragment int
		}
	}
}

func (c *Config) NewPop(popGenerator Operator) *Pop {
	p := New()
	p.Size = c.Size
	p.Length = c.Length
	p.Alphabet = []byte(c.Alphabet)
	popGenerator.Operate(p)

	return p
}

func (c *Config) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Population size: %d\n", c.Size)
	fmt.Fprintf(&b, "Genome length: %d\n", c.Length)
	fmt.Fprintf(&b, "Alphabet: %s\n", string(c.Alphabet))
	fmt.Fprintf(&b, "Mutation rate: %f\n", c.Mutation.Rate)
	fmt.Fprintf(&b, "Transfer rate (in): %f\n", c.Transfer.In.Rate)
	fmt.Fprintf(&b, "Transfer fragment (in): %d\n", c.Transfer.In.Fragment)
	fmt.Fprintf(&b, "Transfer rate (out): %f\n", c.Transfer.Out.Rate)
	fmt.Fprintf(&b, "Transfer fragment (out): %d\n", c.Transfer.Out.Fragment)

	return b.String()
}
