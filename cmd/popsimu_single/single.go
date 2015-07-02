package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/gomath/stat/desc"
	"github.com/mingzhi/gsl-cgo/randist"
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
	ncpu       int
	sampleSize int
	input      string
	output     string
)

func init() {
	const (
		defaultMaxl = 200
	)
	defaultNCPU := runtime.NumCPU()

	flag.IntVar(&ncpu, "ncpu", defaultNCPU, "ncpu")
	flag.IntVar(&sampleSize, "sample", 1000, "sample size of lineages")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)

	runtime.GOMAXPROCS(ncpu)
}

func main() {
	configChan := read(input)
	configMap := make(map[int][]pop.Config)
	for cfg := range configChan {
		configMap[cfg.Length] = append(configMap[cfg.Length], cfg)
	}

	var results []Result
	for seqLen, cfgs := range configMap {
		cfgChan := make(chan pop.Config)
		go func() {
			defer close(cfgChan)
			for _, cfg := range cfgs {
				cfgChan <- cfg
			}
		}()
		res := run(cfgChan, seqLen)
		results = append(results, res...)
	}

	write(output, results)
}

type popConfig struct {
	p *pop.Pop
	c pop.Config
}

func run(configChan chan pop.Config, seqLen int) []Result {
	simResChan := batchSimu(configChan)
	calcChan := calc(simResChan, seqLen)
	results := collect(calcChan)
	return results
}

func batchSimu(configChan chan pop.Config) (resChan chan popConfig) {
	numWorker := runtime.GOMAXPROCS(0)
	resChan = make(chan popConfig, numWorker)
	done := make(chan bool)
	simulator := func() {
		defer send(done)
		for c := range configChan {
			p := simu(c)
			pc := popConfig{p: p, c: c}
			resChan <- pc
		}
	}

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
	ks         *calculator.Ks
	ct         *calculator.AutoCovFFTW
	t2, t3, t4 []float64
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
	c.t2 = append(c.t2, c2.t2...)
	c.t3 = append(c.t3, c2.t3...)
	c.t4 = append(c.t4, c2.t4...)
}

type calcConfig struct {
	cfg pop.Config
	c   *calculators
}

func calc(simResChan chan popConfig, seqLen int) chan calcConfig {
	numWorker := runtime.GOMAXPROCS(0)
	circular := true
	dft := correlation.NewFFTW(seqLen, circular)
	done := make(chan bool)

	calcChan := make(chan calcConfig, numWorker)
	worker := func() {
		defer send(done)
		for res := range simResChan {
			cc := calcConfig{}
			cc.cfg = res.c

			sequences := [][]byte{}
			for _, g := range res.p.Genomes {
				sequences = append(sequences, g.Seq())
			}
			ks := calculator.CalcKs(sequences)
			ct := calculator.CalcCtFFTW(sequences, &dft)
			t2 := pop.CalcT2(res.p, sampleSize)
			t3 := pop.CalcT3(res.p, sampleSize)
			t4 := pop.CalcT4(res.p, sampleSize)

			cc.c = &calculators{}
			cc.c.ks = ks
			cc.c.ct = ct
			cc.c.t2 = t2
			cc.c.t3 = t3
			cc.c.t4 = t4

			calcChan <- cc
		}
	}

	for i := 0; i < numWorker; i++ {
		go worker()
	}

	go func() {
		defer dft.Close()
		defer close(calcChan)
		wait(done, numWorker)
	}()

	return calcChan
}

func collect(calcChan chan calcConfig) []Result {
	m := make(map[pop.Config]*calculators)
	varm := make(map[pop.Config]*desc.Variance)
	for cc := range calcChan {
		c, found := m[cc.cfg]
		v, _ := varm[cc.cfg]
		if !found {
			c = cc.c
			v = desc.NewVariance()
		} else {
			c.Append(cc.c)
		}

		v.Increment(c.ks.Mean.GetResult())

		m[cc.cfg] = c
		varm[cc.cfg] = v
	}

	var results []Result
	for cfg, c := range m {
		res := Result{}
		res.Config = cfg
		res.C = createCovResult(c)
		res.C.KsVar = varm[cfg].GetResult()
		res.T2 = c.t2
		res.T3 = c.t3
		res.T4 = c.t4
		results = append(results, res)
	}

	return results
}

func newPop(c pop.Config, src rand.Source) *pop.Pop {
	p := pop.New()
	r := rand.New(src)
	g := pop.NewRandomPopGenerator(r, c.Size, c.Length, []byte(c.Alphabet))
	g.Operate(p)
	return p
}

func generateEvents(p *pop.Pop, sampler pop.Sampler, mutateEvents []*pop.Event, numGen int, rng *randist.RNG) chan *pop.Event {
	c := make(chan *pop.Event)

	go func() {
		defer close(c)
		mutateRate := 0.0
		for _, e := range mutateEvents {
			mutateRate += e.Rate
		}

		for i := 0; i < numGen; i++ {
			// reproduction.
			samplerEvent := &pop.Event{
				Ops: sampler,
				Pop: p,
			}

			sampler.Start()
			go func() { c <- samplerEvent }()
			sampler.Wait()

			t := sampler.Time(p)
			num := randist.PoissonRandomInt(rng, mutateRate*t*float64(p.Size()))
			for j := 0; j < num; j++ {
				c <- pop.Emit(mutateEvents)
			}
		}
	}()

	return c
}

func simu(c pop.Config) *pop.Pop {
	seed := time.Now().UnixNano()
	src := rand.NewSource(seed)

	p := newPop(c, src)
	r := rand.New(src)

	rng := randist.NewRNG(randist.MT19937_1999)
	defer rng.Free()
	rng.Seed(time.Now().UnixNano())

	var sampler pop.Sampler
	switch c.SampleMethod {
	case "WrightFisher":
		sampler = pop.NewWrightFisherSampler(rng)
	case "LinearSelection":
		sampler = pop.NewLinearSelectionSampler(rng)
	default:
		sampler = pop.NewMoranSampler(rng)
	}

	mutationEvent := &pop.Event{
		Ops:  pop.NewSimpleMutator(r, []byte(c.Alphabet)),
		Pop:  p,
		Rate: c.Mutation.Rate * float64(c.Length),
	}

	fMutator := pop.NewFitnessMutator(c.Mutation.Beneficial.S, 0, rng, pop.MutateStep)
	beneficialMutationEvent := &pop.Event{
		Ops:  fMutator,
		Pop:  p,
		Rate: c.Mutation.Beneficial.Rate * float64(c.Length),
	}

	// choosing fragment size generator.
	var fragGenerator pop.FragSizeGenerator
	switch c.FragGenerator {
	case "exponential":
		lambda := 1.0 / float64(c.Transfer.In.Fragment)
		fragGenerator = pop.NewExpFrag(lambda, src)
	default:
		fragGenerator = pop.NewConstantFrag(c.Transfer.In.Fragment)
	}

	transferEvent := &pop.Event{
		Ops:  pop.NewSimpleTransfer(fragGenerator, r),
		Pop:  p,
		Rate: c.Transfer.In.Rate * float64(c.Length),
	}

	otherEvents := []*pop.Event{mutationEvent, transferEvent, beneficialMutationEvent}
	eventChan := generateEvents(p, sampler, otherEvents, c.NumGen, rng)

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
