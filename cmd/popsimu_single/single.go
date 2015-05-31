package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/gomath/random"
	. "github.com/mingzhi/popsimu/cmd"
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

func main() {
	configChan := read(input)
	results := run(configChan)
	write(output, results)
}

type popConfig struct {
	p *pop.Pop
	c pop.Config
}

func run(configChan chan pop.Config) []Result {
	seed := time.Now().UnixNano()
	src := random.NewLockedSource(rand.NewSource(seed))
	simResChan := batchSimu(configChan, src)
	calcChan := calc(simResChan, maxl)
	results := collect(calcChan)
	return results
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
	ks *calculator.Ks
	ct *calculator.AutoCovFFT
}

func (c *calculators) Increment(xs []float64) {
	for i := 0; i < len(xs); i++ {
		c.ks.Increment(xs[i])
	}

	c.ct.Increment(xs)
}

func (c *calculators) Append(c2 *calculators) {
	c.ks.Append(c2.ks)
	c.ct.Append(c2.ct)
}

func newCalculators(maxl int, circular bool) *calculators {
	c := calculators{}
	c.ks = calculator.NewKs()
	c.ct = calculator.NewAutoCovFFT(maxl, circular)
	return &c
}

type calcConfig struct {
	cfg pop.Config
	c   *calculators
}

func calc(simResChan chan popConfig, maxl int) chan calcConfig {
	circular := true
	done := make(chan bool)

	calcChan := make(chan calcConfig)
	worker := func() {
		defer send(done)
		for res := range simResChan {
			cc := calcConfig{}
			cc.cfg = res.c

			sequences := [][]byte{}
			for _, g := range res.p.Genomes {
				sequences = append(sequences, g.Sequence)
			}
			ks := calculator.CalcKs(sequences)
			ct := calculator.CalcCtFFT(sequences, maxl, circular)

			cc.c = &calculators{}
			cc.c.ks = ks
			cc.c.ct = ct

			calcChan <- cc
		}
	}

	numWorker := runtime.GOMAXPROCS(0)
	for i := 0; i < numWorker; i++ {
		go worker()
	}

	go func() {
		defer close(calcChan)
		wait(done, numWorker)
	}()

	return calcChan
}

func collect(calcChan chan calcConfig) []Result {
	m := make(map[pop.Config]*calculators)
	for cc := range calcChan {
		c, found := m[cc.cfg]
		if !found {
			c = cc.c
		}
		c.Append(cc.c)
		m[cc.cfg] = c
	}

	var results []Result
	for cfg, c := range m {
		res := Result{}
		res.Config = cfg
		res.C = createCovResult(c)
		results = append(results, res)
	}

	return results
}

func newPop(c pop.Config, src rand.Source) *pop.Pop {
	r := rand.New(src)
	popGenerator := pop.NewRandomPopGenerator(r)
	p := c.NewPop(popGenerator)
	return p
}

func simu(c pop.Config, src rand.Source) *pop.Pop {
	p := newPop(c, src)
	r := rand.New(src)

	moranEvent := &pop.Event{
		Ops: pop.NewMoranSampler(r),
		Pop: p,
	}

	mutationEvent := &pop.Event{
		Ops:  pop.NewSimpleMutator(r),
		Pop:  p,
		Rate: c.Mutation.Rate,
	}

	transferEvent := &pop.Event{
		Ops:  pop.NewSimpleTransfer(c.Transfer.In.Fragment, r),
		Pop:  p,
		Rate: c.Transfer.In.Rate,
	}

	// total rate for mutation and transfer
	// for a genome.
	totalRate := (c.Mutation.Rate + c.Transfer.In.Rate) * float64(c.Length)
	poissonSampler := random.NewPoisson(totalRate, src)

	eventChan := make(chan *pop.Event)
	eventGenerator := func() {
		defer close(eventChan)
		events := []*pop.Event{
			mutationEvent,
			transferEvent,
		}
		for i := 0; i < c.NumGen; i++ {
			eventChan <- moranEvent
			num := poissonSampler.Int()
			for j := 0; j < num; j++ {
				eventChan <- pop.Emit(events, r)
			}
		}
	}

	go eventGenerator()

	pop.Evolve(eventChan)

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

func read(filename string) (configChan chan pop.Config) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	configs := readConfigs(f)
	configChan = make(chan pop.Config)

	makeConfigChan := func() {
		defer close(configChan)
		for _, c := range configs {
			configChan <- c
		}
	}

	go makeConfigChan()

	return
}

func readConfigs(r io.Reader) (configs []pop.Config) {
	d := json.NewDecoder(r)
	err := d.Decode(&configs)
	if err != nil {
		panic(err)
	}
	return
}

func createCovResult(c *calculators) CovResult {
	var cr CovResult
	cr.Ks = c.ks.Mean.GetResult()
	for i := 0; i < c.ct.N; i++ {
		cr.Ct = append(cr.Ct, c.ct.GetResult(i))
	}
	return cr
}

func write(filename string, results []Result) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(results); err != nil {
		panic(err)
	}
}
