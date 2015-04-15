package main

import (
	"fmt"
	"github.com/mingzhi/gsl-cgo/randist"
	"github.com/mingzhi/popsimu/pop"
	"os"
	"path/filepath"
)

// This command implements simulation of a single population with horizontal gene transfer.

type cmdSinglePop struct {
	cmdConfig
	rng *randist.RNG // we use gsl random library.
}

// Initialize command.
// It parse flags and configure file settings.
// and invoke config command init function.
func (c *cmdSinglePop) Init() {
	c.Parse()
	c.cmdConfig.Init()

	// initalize random number generator
	c.rng = randist.NewRNG(randist.MT19937_1999)
}

// Run simulations.
func (c *cmdSinglePop) Run(args []string) {
	c.Init()
	ksMV := NewMeanVar()
	vdMV := NewMeanVar()
	for i := 0; i < c.popNum; i++ {
		p := c.RunOne()
		// calcualte population parameters.
		ks, vd := pop.CalcKs(p)
		ksMV.Increment(ks)
		vdMV.Increment(vd)
	}

	outFileName := c.outPrefix + "_ks.txt"
	outFilePath := filepath.Join(c.workspace, c.outDir, outFileName)
	o, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer o.Close()
	o.WriteString("#Ks\tKsVar\tVd\tVdVar\tn\n")
	o.WriteString(fmt.Sprintf("%f\t%f\t%f\t%f\t%d\n", ksMV.Mean.GetResult(), ksMV.Var.GetResult(), vdMV.Mean.GetResult(), vdMV.Var.GetResult(), vdMV.Mean.GetN()))
}

// Run one simulation.
func (c *cmdSinglePop) RunOne() *pop.Pop {
	// initalize population
	p := pop.New()
	p.Size = c.popSize
	p.Length = c.genomeLen
	p.Alphabet = []byte{1, 2, 3, 4}

	rand := randist.NewUniform(c.rng)
	// population operators
	popGenOps := pop.NewRandomPopGenerator(rand)
	moranOps := pop.NewMoranSampler(rand)
	mutationOps := pop.NewSimpleMutator(c.mutRate, rand)
	transferOps := pop.NewSimpleTransfer(c.inTraRate, c.fragSize, rand)

	// initalize the population
	popGenOps.Operate(p)

	// generate operations
	opsChan := make(chan pop.Operator)
	go func() {
		defer close(opsChan)
		for i := 0; i < c.generations; i++ {
			opsChan <- moranOps
			tInt := randist.ExponentialRandomFloat64(c.rng, 1.0/float64(p.Size))
			totalRate := tInt * float64(p.Size*p.Length) * (c.mutRate + c.inTraRate)
			count := randist.PoissonRandomInt(c.rng, totalRate)
			for j := 0; j < count; j++ {
				v := rand.Float64()
				if v <= c.mutRate/(c.mutRate+c.inTraRate) {
					opsChan <- mutationOps
				} else {
					opsChan <- transferOps
				}
			}
		}
	}()

	pop.Evolve(p, opsChan)
	return p
}
