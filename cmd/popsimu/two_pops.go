package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mingzhi/popsimu/pop"
	"github.com/mingzhi/popsimu/simu"
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

	randomSrc := rand.NewSource(time.Now().UnixNano())

	results := Results{PopConfigs: c.popConfigs}
	for i := 0; i < c.numRep; i++ {
		pops := c.RunOne(randomSrc)
		for i := 0; i < len(c.popConfigs); i++ {
			p1 := pops[i]
			ks, vd := pop.CalcKs(c.sampleSize, randomSrc, p1)
			res := CalcRes{
				Index: []int{i},
				Ks:    ks,
				Vd:    vd,
			}

			results.CalcResults = append(results.CalcResults, res)
			for j := i + 1; j < len(c.popConfigs); j++ {
				p2 := pops[j]
				ks, vd := pop.CrossKs(c.sampleSize, randomSrc, p1, p2)
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

func (c *cmdTwoPops) RunOne(src rand.Source) []*pop.Pop {
	// Generate population list.
	pops := make([]*pop.Pop, len(c.popConfigs))
	for i := 0; i < len(pops); i++ {
		pops[i] = newPop(c.popConfigs[i], src)
	}

	simu.Moran(pops, c.popConfigs, c.numGen)

	return pops
}

func newPop(c pop.Config, src rand.Source) *pop.Pop {
	p := pop.New()
	r := rand.New(src)
	g := pop.NewRandomPopGenerator(r, c.Size, c.Length, []byte(c.Alphabet))
	g.Operate(p)
	return p
}
