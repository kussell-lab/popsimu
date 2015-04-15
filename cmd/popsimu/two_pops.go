package main

import (
	"flag"
	"fmt"
	"github.com/jacobstr/confer"
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/popsimu/pop"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
)

type cmdTwoPops struct {
	workspace  string
	configFile string

	numCPU  int
	numGens int
	numReps int

	popConfigs []PopConfig
}

func (c *cmdTwoPops) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&c.workspace, "w", "", "workspace")
	fs.StringVar(&c.configFile, "c", "config.yaml", "configure yaml file")
	fs.IntVar(&c.numCPU, "ncpu", runtime.NumCPU(), "number of CPUs for using")
	return fs
}

func (c *cmdTwoPops) Parse() {
	// Use confer package to parse configure file.
	config := confer.NewConfig()
	config.SetRootPath(c.workspace)
	if err := config.ReadPaths(c.configFile); err != nil {
		panic(err)
	}
	config.AutomaticEnv()

	popFileName := config.GetString("pop.file")
	popFilePath := filepath.Join(c.workspace, popFileName)
	c.popConfigs = parsePopConfigs(popFilePath)

	c.numGens = config.GetInt("pop.generations")
	c.numReps = config.GetInt("pop.replicates")
}

func (c *cmdTwoPops) Init() {
	c.Parse()
	runtime.GOMAXPROCS(c.numCPU)
}

func (c *cmdTwoPops) Run(args []string) {
	c.Init()

	ksMV := NewMeanVar()
	vdMV := NewMeanVar()

	for i := 0; i < c.numReps; i++ {
		c.RunOne()
		// calculate population parameters.
		p1 := c.popConfigs[0].Pop
		p2 := c.popConfigs[1].Pop
		ks, vd := pop.CrossKs(p1, p2)
		ksMV.Increment(ks)
		vdMV.Increment(vd)
	}
	ks := ksMV.Mean.GetResult()
	vd := vdMV.Mean.GetResult()
	fmt.Printf("%f\t%f\n", ks, math.Sqrt(vd)/ks)

	fmt.Printf("%f\t%f\t%f\t%f\t%d\n", ksMV.Mean.GetResult(), ksMV.Var.GetResult(), vdMV.Mean.GetResult(), vdMV.Var.GetResult(), vdMV.Mean.GetN())
}

func (c *cmdTwoPops) RunOne() {
	// Prepare random number generator.
	src := random.NewLockedSource(rand.NewSource(1))

	// Initalize populations.
	for i := 0; i < len(c.popConfigs); i++ {
		c.popConfigs[i].Init()
	}

	// Create commom sequence ancestor.
	seqAncestor := randomGenerateAncestor(c.popConfigs[0].Length, c.popConfigs[0].Alphabet)
	popGenerator := pop.NewSimplePopGenerator(seqAncestor)
	for i := 0; i < len(c.popConfigs); i++ {
		popGenerator.Operate(c.popConfigs[i].Pop)
	}

	ratio := float64(c.popConfigs[0].Pop.Size) / float64(c.popConfigs[0].Pop.Size+c.popConfigs[1].Pop.Size)
	mutRate1, mutRate2 := c.popConfigs[0].Mutation.Rate, c.popConfigs[1].Mutation.Rate
	inTraRate1, inTraRate2 := c.popConfigs[0].Transfer.Rate.In, c.popConfigs[1].Transfer.Rate.In
	outTraRate1, outTraRate2 := c.popConfigs[0].Transfer.Rate.Out, c.popConfigs[1].Transfer.Rate.Out
	frag1, frag2 := c.popConfigs[0].Transfer.Fragment, c.popConfigs[1].Transfer.Fragment
	genomeLen1, genomeLen2 := c.popConfigs[0].Length, c.popConfigs[1].Length

	rates := []float64{
		mutRate1 * ratio * float64(genomeLen1),
		mutRate2 * (1 - ratio) * float64(genomeLen2),
		inTraRate1 * ratio * float64(genomeLen1),
		inTraRate2 * (1 - ratio) * float64(genomeLen2),
		outTraRate1 * ratio * float64(genomeLen1),
		outTraRate2 * (1 - ratio) * float64(genomeLen2),
	}

	totalRate := 0.0
	for _, v := range rates {
		totalRate += v
	}

	poisson := random.NewPoisson(totalRate, src)
	r := rand.New(src)

	moranOps := pop.NewMoranSampler(r)
	mutOps1 := pop.NewSimpleMutator(mutRate1, r)
	mutOps2 := pop.NewSimpleMutator(mutRate2, r)
	inTraOps1 := pop.NewSimpleTransfer(inTraRate1, frag1, r)
	inTraOps2 := pop.NewSimpleTransfer(inTraRate2, frag2, r)
	outTraOps1 := pop.NewOutTransfer(outTraRate1, frag1, c.popConfigs[1].Pop, r)
	outTraOps2 := pop.NewOutTransfer(outTraRate1, frag1, c.popConfigs[0].Pop, r)
	operations := []pop.Operator{
		mutOps1,
		mutOps2,
		inTraOps1,
		inTraOps2,
		outTraOps1,
		outTraOps2,
	}

	go func() {
		for i := 0; i < c.numGens; i++ {
			popIndex := findCategory([]float64{
				float64(c.popConfigs[0].Pop.Size),
				float64(c.popConfigs[1].Pop.Size),
			}, r)
			c.popConfigs[popIndex].Pop.OpsChan <- moranOps

			count := poisson.Int()
			for j := 0; j < count; j++ {
				opsIndex := findCategory(rates, r)
				popIndex := opsIndex % 2
				c.popConfigs[popIndex].Pop.OpsChan <- operations[opsIndex]
			}
		}

		for _, pc := range c.popConfigs {
			close(pc.Pop.OpsChan)
		}
	}()

	done := make(chan bool)
	for i := 0; i < 2; i++ {
		go func(index int) {
			c.popConfigs[index].Pop.Evolve()
			done <- true
		}(i)
	}

	for i := 0; i < 2; i++ {
		<-done
	}
}

func findCategory(proportions []float64, r *rand.Rand) (index int) {
	totalRate := 0.0
	for _, v := range proportions {
		totalRate += v
	}

	rv := r.Float64()

	t := 0.0
	for i := 0; i < len(proportions); i++ {
		t += proportions[i] / totalRate
		if rv <= t {
			index = i
			return
		}
	}

	return
}

type PopConfig struct {
	Pop *pop.Pop

	Size     int
	Length   int
	Alphabet []byte
	Mutation struct {
		Rate float64
	}

	Transfer struct {
		Rate struct {
			In, Out float64
		}
		Fragment int
	}
}

func (p *PopConfig) Init() {
	p.Alphabet = []byte{1, 2, 3, 4}
	p.Pop = pop.New()
	p.Pop.Size = p.Size
	p.Pop.Length = p.Length
	p.Pop.Alphabet = p.Alphabet
	p.Pop.OpsChan = make(chan pop.Operator)
}

func parsePopConfigs(filename string) (popCfgs []PopConfig) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(content, &popCfgs); err != nil {
		panic(err)
	}

	return
}

func randomGenerateAncestor(size int, alphbets []byte) pop.Sequence {
	s := make(pop.Sequence, size)
	for i := 0; i < size; i++ {
		s[i] = alphbets[rand.Intn(len(alphbets))]
	}
	return s
}
