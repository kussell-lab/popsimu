package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"fmt"

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

// CalcRes stores calculation results.
type CalcRes struct {
	ID string
	Ks float64
	Vd float64
	Cs []float64
	Ct []float64
	Cm []float64
	Cr []float64
}

func (c *cmdTwoPops) Run(args []string) {
	c.ParsePopConfigs()
	for i := 0; i < len(c.popConfigs); i++ {
		c.popConfigs[i].NumGen = c.numGen
	}

	runtime.GOMAXPROCS(c.ncpu)
	resChan := make(chan Results)
	done := make(chan bool)
	for i := 0; i < c.numRep; i++ {
		go func() {
			results := Results{PopConfigs: c.popConfigs}
			randomSrc := rand.NewSource(time.Now().UnixNano())
			pops := c.RunOne(randomSrc)
			for i := 0; i < len(c.popConfigs); i++ {
				p1 := pops[i]
				ks, vd := pop.CalcKs(c.sampleSize, randomSrc, p1)
				cm, ct, cr, cs := pop.CalcCov(c.sampleSize, c.maxl, randomSrc, p1)
				res := CalcRes{
					ID: fmt.Sprintf("%d", i),
					Ks: ks,
					Vd: vd,
					Cm: cm,
					Ct: ct,
					Cr: cr,
					Cs: cs,
				}
				results.CalcResults = append(results.CalcResults, res)
				for j := i + 1; j < len(c.popConfigs); j++ {
					p2 := pops[j]
					ks, vd := pop.CrossKs(c.sampleSize, randomSrc, p1, p2)
					cm, ct, cr, cs := pop.CrossCov(c.sampleSize, c.maxl, randomSrc, p1, p2)
					res := CalcRes{
						ID: fmt.Sprintf("%d_%d", i, j),
						Ks: ks,
						Vd: vd,
						Cm: cm,
						Ct: ct,
						Cr: cr,
						Cs: cs,
					}
					results.CalcResults = append(results.CalcResults, res)
				}
			}
			resChan <- results
			done <- true
		}()
	}

	go func() {
		defer close(resChan)
		for i := 0; i < c.numRep; i++ {
			<-done
		}
	}()

	resMap := collect(resChan)

	outFileName := c.prefix + "_res.csv"
	outFilePath := filepath.Join(c.outdir, outFileName)
	w, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("l,m,v,n,t,b\n")
	for key, values := range resMap {
		results := values.Results()
		for _, res := range results {
			w.WriteString(fmt.Sprintf("%d,%g,%g,%d,%s,%s\n", res.Lag, res.Mean, res.Variance, res.N, res.Type, key))
		}
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

func collect(resChan chan Results) map[string]*Collector {
	resMap := make(map[string]*Collector)
	for resCfg := range resChan {
		for _, res := range resCfg.CalcResults {
			key := res.ID
			if _, found := resMap[key]; !found {
				resMap[key] = NewCollector()
			}

			var corrResults CorrResults
			for i := range res.Ct {
				var corrRes CorrResult
				corrRes.Lag = i
				corrRes.Mean = res.Ct[i]
				corrRes.Type = "P2"
				corrResults.Results = append(corrResults.Results, corrRes)
			}
			resMap[key].Add(corrResults)
		}
	}
	return resMap
}
