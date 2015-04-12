package main

import (
	"fmt"
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/gomath/stat/desc"
	"github.com/mingzhi/popsimu/pop"
	"math"
	"math/rand"
	"time"
)

// This command implements simulation of a single population with horizontal gene transfer.

type cmdSinglePop struct {
	cmdConfig
}

func (c *cmdSinglePop) Init() {
	c.Parse()
	c.cmdConfig.Init()
}

func (c *cmdSinglePop) Run(args []string) {
	c.Init()
	mean := desc.NewMean()
	variance := desc.NewVarianceWithBiasCorrection()
	for i := 0; i < c.popNum; i++ {
		p := c.RunOne()
		// calculator
		ksCalc := pop.NewKsCalculator()
		v := ksCalc.Calc(p)
		mean.Increment(v)
		variance.Increment(v)
	}

	nu := float64(c.popSize) * c.mutRate
	expect := nu / (1 + 4.0/3.0*nu)
	result := mean.GetResult()
	stderr := math.Sqrt(variance.GetResult() / float64(variance.GetN()))
	fmt.Println(expect)
	fmt.Printf("%f - %f\n", result-stderr, result+stderr)
}

func (c *cmdSinglePop) RunOne() *pop.Pop {
	// random number generator.
	seed := time.Now().UnixNano()
	src := rand.NewSource(seed)
	r := rand.New(src)
	// initalize population.
	p := pop.New()
	p.Size = c.popSize
	p.Length = c.genomeLen
	p.Alphabet = []byte{'A', 'T', 'G', 'C'}
	// randomly generate the genome sequences.
	popGenerator := pop.NewRandomPopGenerator(r)
	popGenerator.Operate(p)
	// possion generator.
	t := 1.0 / float64(p.Size)
	rate := t * (c.mutRate + c.inTraRate) * float64(p.Length*p.Size)
	poisson := random.NewPoisson(1.0/rate, src)
	// operators
	sampler := pop.NewMoranSampler(r)
	mutator := pop.NewSimpleMutator(c.mutRate, r)
	transfer := pop.NewSimpleTransfer(c.inTraRate, c.fragSize, r)
	for i := 0; i < c.generations; i++ {
		sampler.Operate(p)
		count := poisson.Int()
		for j := 0; j < count; j++ {
			// determine if it is a mutation or a transfer.
			v := r.Float64()
			if v <= c.mutRate/(c.mutRate+c.inTraRate) {
				mutator.Operate(p)
			} else {
				transfer.Operate(p)
			}
		}

	}

	return p
}
