package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mingzhi/popsimu/pop"
	"os"
)

var (
	input  string
	output string
	rep    int
)

func main() {
	flag.IntVar(&rep, "r", 1, "replicates")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)
	ps := parse(input)
	configs := create(ps)
	write(output, configs)
	fmt.Println(ps)
}

type ParameterSet struct {
	Sizes                   []int
	Lengths                 []int
	MutationRates           []float64
	TransferInRates         []float64
	TransferInFrags         []int
	TransferOutRates        []float64
	TransferOutFrags        []int
	Alphabet                string
	BeneficialMutationRates []float64
	FitnessEffects          []float64
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
	fmt.Fprintf(&b, "Alphabet: %v\n", p.Alphabet)
	fmt.Fprintf(&b, "Beneficial mutation rates: %v\n", p.BeneficialMutationRates)
	fmt.Fprintf(&b, "Fitness Effect: %v\n", p.FitnessEffects)

	return b.String()
}

func parse(filePath string) ParameterSet {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d := json.NewDecoder(f)

	sets := ParameterSet{}
	if err := d.Decode(&sets); err != nil {
		panic(err)
	}
	return sets
}

func create(par ParameterSet) []pop.Config {
	cfgs := []pop.Config{}
	for _, size := range par.Sizes {
		for _, length := range par.Lengths {
			for _, mutation := range par.MutationRates {
				for _, transferInRate := range par.TransferInRates {
					for _, transferInFrag := range par.TransferInFrags {
						for _, transferOutRate := range par.TransferOutRates {
							for _, transferOutFrag := range par.TransferOutFrags {
								for _, beneficalMutationRate := range par.BeneficialMutationRates {
									for _, s := range par.FitnessEffects {
										for i := 0; i < rep; i++ {
											cfg := pop.Config{}
											cfg.Size = size
											cfg.Length = length
											cfg.Mutation.Rate = mutation
											cfg.Transfer.In.Rate = transferInRate
											cfg.Transfer.In.Fragment = transferInFrag
											cfg.Transfer.Out.Rate = transferOutRate
											cfg.Transfer.Out.Fragment = transferOutFrag
											cfg.Alphabet = par.Alphabet
											cfg.Mutation.Beneficial.Rate = beneficalMutationRate
											cfg.Mutation.Beneficial.S = s
											cfg.NumGen = cfg.Size * cfg.Size * 10
											cfgs = append(cfgs, cfg)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return cfgs
}

func write(filename string, configs []pop.Config) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	e := json.NewEncoder(w)
	if err := e.Encode(configs); err != nil {
		panic(err)
	}
}
