package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/popsimu/pop"
	"github.com/mingzhi/seqcor/calculator"
	"io"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var (
	maxl   int
	ncpu   int
	input  string
	output string
)

func init() {
	const (
		defaultMaxl = 200
	)
	defaultNCPU := runtime.NumCPU()

	flag.IntVar(&maxl, "maxl", defaultMaxl, "maxl")
	flag.IntVar(&ncpu, "ncpu", defaultNCPU, "ncpu")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)

	runtime.GOMAXPROCS(ncpu)
}

type popConfig struct {
	p *pop.Pop
	c pop.Config
}

func run(configChan chan Config) {
	seed := time.Now().UnixNano()
	src := random.NewLockedSource(rand.NewSource(seed))
	simResChan := batchSimu(configChan, src)
}

func batchSimu(configChan chan pop.Config, src rand.Source) (resChan chan popConfig) {
	resChan = make(chan popConfig)
	done := make(chan bool)
	simulator := func() {
		defer send(done)
		for c := range configChan {
			p := simu(c, src)
			pc := popConfig{p: p, c: c}
			resChan <- pc
		}
	}

	numWorker := runtime.GOMAXPROCS(0)
	for i := 0; i < numWorker; i++ {
		go simulator()
	}

	go func() {
		defer close(resChan)
		wait(done, numWorker)
	}()

	return resChan
}

type calculators struct {
	ks *calculator.KsCalculator
	ct *calculator.CovCalculatorFFT
}

func (c *calculators) Increment(xs []float64) {
	for i := 0; i < len(xs); i++ {
		c.ks.Increment(xs[i])
	}

	c.ct.Increment(xs)
}

func newCalculators(maxl int, circular bool) *calculators {
	c := calculators{}
	c.ks = NewKsCalculator()
	c.ct = NewCovCalculatorFFT(maxl, circular)
	return &c
}

func calc(simResChan chan popConfig, maxl int) {
	circular := true
	done := make(chan bool)
	calculator := func() {
		defer send(done)
		c := newCalculators(maxl, circular)
		for res := range simResChan {
			sequences := [][]byte{}
			for _, g := range res.p.Genomes {
				sequences = append(sequences, g.Sequence)
			}

		}
	}
}

func newPop(c Config, src rand.Source) *Pop {
	r := rand.New(src)
	popGenerator := NewRandomPopGenerator(r)
	p := c.NewPop(popGenerator)
	return p
}

func simu(c Config, src rand.Source) *Pop {
	p := newPop(c, src)
	r := rand.New(src)

	moranEvent := &Event{
		Ops: NewMoranSampler(r),
		Pop: p,
	}

	mutationEvent := &Event{
		Ops:  NewSimpleMutator(r),
		Pop:  p,
		Rate: c.Mutation.Rate,
	}

	transferEvent := &Event{
		Ops:  NewSimpleTransfer(c.Transfer.In.Fragment, r),
		Pop:  p,
		Rate: c.Transfer.In.Rate,
	}

	// total rate for mutation and transfer
	// for a genome.
	totalRate := (c.Mutation.Rate + c.Transfer.In.Rate) * float64(c.Length)
	poissonSampler := random.NewPoisson(totalRate, src)

	eventChan := make(chan *Event)
	eventGenerator := func() {
		defer close(eventChan)
		events := []*Event{
			mutationEvent,
			transferEvent,
		}
		for i := 0; i < c.NumGen; i++ {
			eventChan <- moranEvent
			num := poissonSampler.Int()
			for j := 0; j < num; j++ {
				eventChan <- Emit(events, r)
			}
		}
	}

	go eventGenerator()

	Evolve(eventChan)

	return p
}

func send(done chan bool) {
	done <- true
}

func wait(done chan bool, numWorker int) {
	for i := 0; i < numWorker; i++ {
		<-done
	}
}

func read(filename string) (configChan chan Config) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	configs := readConfigs(f)
	configChan = make(chan Config)

	makeConfigChan := func() {
		defer close(configChan)
		for _, c := range configs {
			configChan <- c
		}
	}

	go makeConfigChan()

	return
}

func readConfigs(r io.Reader) (configs []Config) {
	d := json.NewDecoder(r)
	err := d.Decode(&configs)
	if err != nil {
		panic(err)
	}
	return
}
