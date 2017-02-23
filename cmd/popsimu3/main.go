package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"math/rand"
	"time"

	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/cheggaaa/pb"
	"github.com/mingzhi/popsimu/pop"
	"github.com/mingzhi/popsimu/simu"
)

func main() {
	app := kingpin.New("popsimu3", "population simulations")
	app.Version("v0.3")
	configFile := app.Arg("config", "population config file").Required().String()
	outFile := app.Arg("output", "output file").Required().String()
	replicates := app.Flag("replicate", "number of replications").Default("1").Int()
	numGen := app.Flag("generation", "number of generation").Default("0").Int()
	ncpu := app.Flag("ncpu", "number of CPUs").Default("0").Int()
	sampleSize := app.Flag("sample_size", "sample size").Default("1000").Int()
	sampleStep := app.Flag("sample_step", "sample step").Default("100").Int()
	sampleTime := app.Flag("sample_time", "sample time").Default("100").Int()
	maxl := app.Flag("maxl", "maxl").Default("100").Int()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *ncpu == 0 {
		*ncpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*ncpu)

	pc := parsePopConfig(*configFile)
	pp := generatePopulations(pc, *replicates)

	if *numGen == 0 {
		*numGen = pc.Size * pc.Size * 10
	}
	if *sampleStep == 0 {
		*sampleStep = pc.Size
	}

	var ancestors []Community
	for _, p := range pp {
		ancestors = append(ancestors, Community{p})
	}
	log.Println("Evolving ancestors...")
	evolute(ancestors, []pop.Config{pc}, *numGen, *ncpu)
	resChan := calculate(ancestors, *sampleSize, *maxl, *ncpu)

	w, err := os.Create(*outFile)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("l,m,v,n,t,b,g\n")
	write(w, resChan, *maxl, "-1")

	var communities []Community
	dilution := pop.Dilution{Factor: 0.1}
	for _, p := range pp {
		p1 := dilution.Reduce(p)
		pop.Recover(p1, p.Size())
		p2 := dilution.Reduce(p)
		pop.Recover(p2, p.Size())
		communities = append(communities, Community{p1, p2})
	}

	for i := 0; i <= *sampleTime; i++ {
		log.Printf("Evolving %d ...\n", i)
		key := fmt.Sprintf("%d", i)
		if i > 0 {
			evolute(communities, []pop.Config{pc, pc}, *numGen, *ncpu)
		}
		resChan2 := calculate(communities, *sampleSize, *maxl, *ncpu)
		write(w, resChan2, *maxl, key)
	}

}

// Community is a group of populations.
type Community []*pop.Pop

// parsePopConfig parse a JSON PopConfig
func parsePopConfig(file string) (pc pop.Config) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&pc); err != nil {
		log.Fatalln(err)
	}
	return
}

func generatePopulations(pc pop.Config, num int) []*pop.Pop {
	var pp []*pop.Pop
	for i := 0; i < num; i++ {
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		p := pop.New()
		g := pop.NewRandomPopGenerator(r, pc.Size, pc.Length, []byte(pc.Alphabet))
		g.Operate(p)
		pp = append(pp, p)
	}
	return pp
}

func evolute(cc []Community, pcList []pop.Config, numGen, ncpu int) {
	cChan := make(chan Community)
	go func() {
		defer close(cChan)
		for _, c := range cc {
			cChan <- c
		}
	}()

	done := make(chan bool)
	for i := 0; i < ncpu; i++ {
		go func() {
			for c := range cChan {
				simu.Moran(c, pcList, numGen)
				done <- true
			}

		}()
	}

	pbar := pb.StartNew(len(cc))
	defer pbar.FinishPrint("Finish evolution.")
	for i := 0; i < len(cc); i++ {
		<-done
		pbar.Increment()
	}

	return
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

func calculate(ppList []Community, sampleSize, maxl, ncpu int) chan CalcRes {
	jobChan := make(chan []*pop.Pop)
	go func() {
		defer close(jobChan)
		for _, pp := range ppList {
			jobChan <- pp
		}
	}()

	resChan := make(chan CalcRes)
	done := make(chan bool)
	for i := 0; i < ncpu; i++ {
		go func() {
			for pp := range jobChan {
				src := rand.NewSource(time.Now().UnixNano())
				for k := 0; k < len(pp); k++ {
					p1 := pp[k]
					ks, vd := pop.CalcKs(sampleSize, src, p1)
					cm, ct, cr, cs := pop.CalcCov(sampleSize, maxl, src, p1)
					res := CalcRes{
						ID: fmt.Sprintf("%d", k),
						Ks: ks,
						Vd: vd,
						Cm: cm,
						Ct: ct,
						Cr: cr,
						Cs: cs,
					}
					resChan <- res
					for j := k + 1; j < len(pp); j++ {
						p2 := pp[j]
						ks, vd := pop.CrossKs(sampleSize, src, p1, p2)
						cm, ct, cr, cs := pop.CrossCov(sampleSize, maxl, src, p1, p2)
						res := CalcRes{
							ID: fmt.Sprintf("%d_%d", k, j),
							Ks: ks,
							Vd: vd,
							Cm: cm,
							Ct: ct,
							Cr: cr,
							Cs: cs,
						}
						resChan <- res
					}
				}
				done <- true
			}

		}()
	}

	go func() {
		pbar := pb.StartNew(len(ppList))
		defer pbar.FinishPrint("Finish calculation.")
		defer close(resChan)
		for i := 0; i < len(ppList); i++ {
			<-done
			pbar.Increment()
		}
	}()

	return resChan
}

func collect(resChan chan CalcRes) map[string]*Collector {
	resMap := make(map[string]*Collector)
	for res := range resChan {
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
			if corrRes.Lag > 0 {
				corrRes.Mean = corrRes.Mean / res.Ks
			}
			corrResults.Results = append(corrResults.Results, corrRes)
		}
		resMap[key].Add(corrResults)
	}
	return resMap
}

func write(w *os.File, resChan chan CalcRes, maxl int, t string) {
	resMap := collect(resChan)

	for key, values := range resMap {
		results := values.Results()
		for _, res := range results {
			if res.Lag <= maxl {
				w.WriteString(fmt.Sprintf("%d,%g,%g,%d,%s,%s,%s\n", res.Lag, res.Mean, res.Variance, res.N, res.Type, key, t))
			}
		}
	}
}
