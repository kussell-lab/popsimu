package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mingzhi/popsimu/pop"
	"github.com/mingzhi/popsimu/simu"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	workspace  string // workspace
	config     string // config file
	prefix     string // prefix
	outdir     string // output folder
	numRep     int    // number of replicates.
	sampleStep int    // number of generations for each step
	sampleTime int    // number of times
	sampleSize int    // sample size
	maxL       int    // max length of correlation
)

func init() {
	// flags
	flag.StringVar(&workspace, "w", "", "workspace")
	flag.StringVar(&config, "c", "config.yaml", "parameter set file")
	flag.StringVar(&prefix, "p", "test", "prefix")
	flag.StringVar(&outdir, "o", "", "output directory")
	flag.IntVar(&numRep, "n", 10, "number of replicates")
	flag.IntVar(&sampleTime, "s", 1000, "number of generations for each step")
	flag.IntVar(&sampleStep, "t", 1, "number of times")
	flag.IntVar(&sampleSize, "sample", 1000, "sample size")
	flag.IntVar(&maxL, "maxl", 100, "max length of correlation")

	flag.Parse()
}

func main() {
	ncpu := runtime.NumCPU()
	runtime.GOMAXPROCS(ncpu)
	// parse parameter sets.
	filePath := filepath.Join(workspace, config)
	fmt.Printf("Config file path: %s\n", filePath)
	par := parseParameterSets(filePath)
	popConfigs := createPopConfigs(par)

	fmt.Println(par)
	fmt.Printf("Total %d population configs.\n", len(popConfigs))

	jobs := make(chan pop.Config)
	results := make(chan Results)
	done := make(chan bool)
	numWorker := ncpu
	// define worker.
	worker := func() {
		for cfg := range jobs {
			res := Results{PopConfigs: []pop.Config{cfg}}
			p := createPop(cfg)
			numGen := 10 * p.Size * p.Size
			t0 := time.Now()
			simu.RunMoran([]*pop.Pop{p}, []pop.Config{cfg}, numGen)
			calcResults := calculateResults([]*pop.Pop{p}, numGen)
			res.CalcResults = append(res.CalcResults, calcResults...)
			t1 := time.Now()
			fmt.Printf("Using %.3f seconds!\n", t1.Sub(t0).Seconds())
			for j := 0; j < numRep; j++ {
				fmt.Printf("Replicate %d\n", j)
				fmt.Println("Splitting the population into two.")

				replacement := true
				pops := splitPop(p, replacement, p.Size, p.Size)
				cfgs := []pop.Config{cfg, cfg}
				numGen1 := numGen
				t0 := time.Now()
				for i := 0; i < sampleStep; i++ {
					simu.RunMoran(pops, cfgs, sampleTime)
					numGen1 += sampleTime
					calcResults := calculateResults(pops, numGen1)
					fmt.Printf("%f, %f, %f\n", calcResults[0].Ks, calcResults[1].Ks, calcResults[3].Ks)
					res.CalcResults = append(res.CalcResults, calcResults...)
				}
				t1 := time.Now()
				fmt.Printf("Evolve %d generations, using %.3f seconds!\n", sampleTime*sampleStep, t1.Sub(t0).Seconds())
			}

			results <- res
		}
		done <- true
	}

	go func() {
		defer close(jobs)
		for _, cfg := range popConfigs {
			jobs <- cfg
		}
	}()

	for i := 0; i < numWorker; i++ {
		go worker()
	}

	go func() {
		defer close(results)
		for i := 0; i < numWorker; i++ {
			<-done
		}
	}()

	finalResults := []Results{}
	for res := range results {
		finalResults = append(finalResults, res)
	}

	outFileName := prefix + "_res.json"
	outFilePath := filepath.Join(workspace, outdir, outFileName)
	outfile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	encoder := json.NewEncoder(outfile)
	if err := encoder.Encode(finalResults); err != nil {
		panic(err)
	}

	fmt.Printf("Finished! Save results to %s\n", outFilePath)
}

type ParameterSet struct {
	Sizes            []int
	Lengths          []int
	MutationRates    []float64
	TransferInRates  []float64
	TransferInFrags  []int
	TransferOutRates []float64
	TransferOutFrags []int
	Alphabet         []byte
}

func (p ParameterSet) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Sizes: %v\n", p.Sizes)
	fmt.Fprintf(&b, "Lengths: %v\n", p.Lengths)
	fmt.Fprintf(&b, "Mutation rates: %v\n", p.MutationRates)
	fmt.Fprintf(&b, "Transfer in rates: %v\n", p.TransferInRates)
	fmt.Fprintf(&b, "Transfer in frags: %v\n", p.TransferInFrags)
	fmt.Fprintf(&b, "Transfer out rates: %v\n", p.TransferOutRates)
	fmt.Fprintf(&b, "Transfer out frags: %v\n", p.TransferOutFrags)

	return b.String()
}

func parseParameterSets(filePath string) ParameterSet {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)

	sets := ParameterSet{}
	if err := yaml.Unmarshal(data, &sets); err != nil {
		panic(err)
	}
	return sets
}

func createPopConfigs(par ParameterSet) []pop.Config {
	cfgs := []pop.Config{}
	for _, size := range par.Sizes {
		for _, length := range par.Lengths {
			for _, mutation := range par.MutationRates {
				for _, transferInRate := range par.TransferInRates {
					for _, transferInFrag := range par.TransferInFrags {
						for _, transferOutRate := range par.TransferOutRates {
							for _, transferOutFrag := range par.TransferOutFrags {
								cfg := pop.Config{}
								cfg.Size = size
								cfg.Length = length
								cfg.Mutation.Rate = mutation
								cfg.Transfer.In.Rate = transferInRate
								cfg.Transfer.In.Fragment = transferInFrag
								cfg.Transfer.Out.Rate = transferOutRate
								cfg.Transfer.Out.Fragment = transferOutFrag
								cfg.Alphabet = par.Alphabet
								cfgs = append(cfgs, cfg)
							}
						}
					}
				}
			}
		}
	}

	return cfgs
}

func createPop(cfg pop.Config) *pop.Pop {
	size := cfg.Length
	alphabet := cfg.Alphabet
	ancestor := randomGenerateAncestor(size, alphabet)
	generator := pop.NewSimplePopGenerator(ancestor)
	return cfg.NewPop(generator)
}

func splitPop(p *pop.Pop, replacement bool, sizes ...int) []*pop.Pop {
	totalSizes := sizes
	pops := []*pop.Pop{}
	if replacement {
		for _, size := range totalSizes {
			p1 := pop.New()
			p1.Size = size
			p1.Length = p.Length
			p1.Alphabet = p.Alphabet
			p1.Circled = p.Circled
			for i := 0; i < p1.Size; i++ {
				index := rand.Intn(p.Size)
				p1.Genomes = append(p1.Genomes, p.Genomes[index])
			}
			pops = append(pops, p1)
		}
	} else {
		indices := make([]int, p.Size)
		for i := 0; i < len(indices); i++ {
			indices[i] = i
		}
		for i := 0; i < len(indices); i++ {
			j := rand.Intn(len(indices))
			indices[i], indices[j] = indices[j], indices[i]
		}

		index := 0
		for _, size := range totalSizes {
			if index+size >= p.Size {
				panic("total sizes bigger than the population size")
			}
			p1 := pop.New()
			p1.Length = p.Length
			p1.Alphabet = p.Alphabet
			p1.Circled = p.Circled
			for i := 0; i < size; i++ {
				p1.Genomes = append(p1.Genomes, p.Genomes[index+i])
			}
			pops = append(pops, p1)
			index += size
		}
	}

	return pops
}

func randomGenerateAncestor(size int, alphbets []byte) pop.Sequence {
	s := make(pop.Sequence, size)
	for i := 0; i < size; i++ {
		s[i] = alphbets[rand.Intn(len(alphbets))]
	}
	return s
}

type Results struct {
	PopConfigs  []pop.Config
	CalcResults []CalcRes
}

type CalcRes struct {
	Index          []int
	Ks, Vd         float64
	Cm, Ct, Cr, Cs []float64
	NumGen         int
}

func calculateResults(pops []*pop.Pop, numGen int) []CalcRes {
	calcResults := []CalcRes{}
	for i := 0; i < len(pops); i++ {
		p1 := pops[i]
		ks, vd := pop.CalcKs(sampleSize, p1)
		res := CalcRes{
			Index:  []int{i},
			Ks:     ks,
			Vd:     vd,
			NumGen: numGen,
		}
		res.Cm, res.Ct, res.Cr, res.Cs = pop.CalcCov(sampleSize, maxL, p1)
		calcResults = append(calcResults, res)

		for j := i + 1; j < len(pops); j++ {
			p2 := pops[j]
			ks, vd := pop.CrossKs(sampleSize, p1, p2)
			res := CalcRes{
				Index:  []int{i, j},
				Ks:     ks,
				Vd:     vd,
				NumGen: numGen,
			}
			res.Cm, res.Ct, res.Cr, res.Cs = pop.CrossCov(sampleSize, maxL, p1, p2)
			calcResults = append(calcResults, res)
		}
	}

	if len(pops) > 1 {
		indices := []int{}
		for i := 0; i < len(pops); i++ {
			indices = append(indices, i)
		}

		p1 := pops[0]
		others := pops[1:]
		ks, vd := pop.CalcKs(sampleSize, p1, others...)
		res := CalcRes{
			Index:  indices,
			Ks:     ks,
			Vd:     vd,
			NumGen: numGen,
		}
		res.Cm, res.Ct, res.Cr, res.Cs = pop.CalcCov(sampleSize, maxL, p1, others...)
		calcResults = append(calcResults, res)
	}

	return calcResults
}
