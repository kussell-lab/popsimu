package main

import (
	"fmt"
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/popsimu/pop"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// This command implements simulation of a single population with horizontal gene transfer.

type cmdSinglePop struct {
	cmdConfig
}

// Initialize command.
// It parse flags and configure file settings.
// and invoke config command init function.
func (c *cmdSinglePop) Init() {
	c.Parse()
	c.cmdConfig.Init()
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
	src := random.NewLockedSource(rand.NewSource(time.Now().UnixNano()))
	r := rand.New(src)
	// initalize population
	p := pop.New()
	p.Size = c.popSize
	p.Length = c.genomeLen
	p.Alphabet = []byte{1, 2, 3, 4}

	// population operators
	popGenOps := pop.NewRandomPopGenerator(r)
	moranOps := pop.NewMoranSampler(r)
	mutationOps := pop.NewSimpleMutator(c.mutRate, r)
	transferOps := pop.NewSimpleTransfer(c.inTraRate, c.fragSize, r)

	totalRate := float64(p.Length) * (c.mutRate + c.inTraRate)
	poisson := random.NewPoisson(totalRate, src)

	// initalize the population
	popGenOps.Operate(p)

	// generate operations
	opsChan := make(chan pop.Operator)
	go func() {
		defer close(opsChan)
		for i := 0; i < c.generations; i++ {
			opsChan <- moranOps
			count := poisson.Int()
			for j := 0; j < count; j++ {
				v := r.Float64()
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
