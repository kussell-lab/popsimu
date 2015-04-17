package main

import (
	"encoding/json"
	"github.com/mingzhi/popsimu/pop"
	"github.com/mingzhi/popsimu/simu"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
)

type cmdTwoPops struct {
	cmdConfig
}

type Results struct {
	PopConfigs  []pop.Config
	CalcResults []CalcRes
}

type CalcRes struct {
	Index []int
	Ks    float64
	Vd    float64
}

func (c *cmdTwoPops) Run(args []string) {
	c.ParsePopConfigs()
	for i := 0; i < len(c.popConfigs); i++ {
		c.popConfigs[i].NumGen = c.numGen
	}

	runtime.GOMAXPROCS(c.ncpu)

	results := Results{PopConfigs: c.popConfigs}
	for i := 0; i < c.numRep; i++ {
		pops := c.RunOne()
		for i := 0; i < len(c.popConfigs); i++ {
			p1 := pops[i]
			ks, vd := pop.CalcKs(p1)
			res := CalcRes{
				Index: []int{i},
				Ks:    ks,
				Vd:    vd,
			}

			results.CalcResults = append(results.CalcResults, res)
			for j := i + 1; j < len(c.popConfigs); j++ {
				p2 := pops[j]
				ks, vd := pop.CrossKs(p1, p2)
				res := CalcRes{
					Index: []int{i, j},
					Ks:    ks,
					Vd:    vd,
				}
				results.CalcResults = append(results.CalcResults, res)
			}
		}
	}

	outFileName := c.prefix + "_res.json"
	outFilePath := filepath.Join(c.outdir, outFileName)
	o, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer o.Close()

	encoder := json.NewEncoder(o)
	if err := encoder.Encode(results); err != nil {
		panic(err)
	}
}

func (c *cmdTwoPops) RunOne() []*pop.Pop {
	// Create population generator (with common ancestor).
	genomeSize := c.popConfigs[0].Length
	alphabet := c.popConfigs[0].Alphabet
	seqAncestor := randomGenerateAncestor(genomeSize, alphabet)
	popGenerator := pop.NewSimplePopGenerator(seqAncestor)

	// Generate population list.
	pops := make([]*pop.Pop, len(c.popConfigs))
	for i := 0; i < len(pops); i++ {
		pops[i] = c.popConfigs[i].NewPop(popGenerator)
	}

	simu.RunMoran(pops, c.popConfigs, c.numGen)

	return pops
}

func randomGenerateAncestor(size int, alphbets []byte) pop.Sequence {
	s := make(pop.Sequence, size)
	for i := 0; i < size; i++ {
		s[i] = alphbets[rand.Intn(len(alphbets))]
	}
	return s
}
