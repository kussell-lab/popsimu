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
		Beneficial struct {
			Rate float64
			S    float64
		}
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

	SampleMethod  string
	FragGenerator string
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
