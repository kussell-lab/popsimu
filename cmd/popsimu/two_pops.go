package main

import (
	"encoding/json"
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/popsimu/pop"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
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
	// Random source.
	randomSeed := time.Now().UnixNano()
	randomSource := random.NewLockedSource(rand.NewSource(randomSeed))

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

	// Prepare possible events.
	events := generateEvents(c.popConfigs, pops, randomSource)
	moranEvents := generateMoranEvents(c.popConfigs, pops, randomSource)

	totalPopSize := 0
	for i := 0; i < len(c.popConfigs); i++ {
		totalPopSize += c.popConfigs[i].Size
	}

	totalRate := 0.0
	for i := 0; i < len(events); i++ {
		// the rate unit is per genome per generation,
		// so we need to rescale it by dividing the population size.
		totalRate += events[i].Rate / float64(totalPopSize)
	}

	// Poisson random source.
	poisson := random.NewPoisson(totalRate, randomSource)
	r := rand.New(randomSource)

	// Create and load event channel.
	eventChan := make(chan *pop.Event)
	go func() {
		defer close(eventChan)
		for i := 0; i < c.numGen; i++ {
			moranEvent := pop.Emit(moranEvents, r)
			eventChan <- moranEvent

			count := poisson.Int()
			for c := 0; c < count; c++ {
				event := pop.Emit(events, r)
				eventChan <- event
			}
		}
	}()

	// Evolve.
	pop.Evolve(eventChan)

	return pops
}

func generateMoranEvents(popConfigs []pop.Config, pops []*pop.Pop, src rand.Source) (moranEvents []*pop.Event) {
	r := rand.New(src)
	for i := 0; i < len(popConfigs); i++ {
		event := &pop.Event{
			Rate: float64(popConfigs[i].Size),
			Ops:  pop.NewMoranSampler(r),
			Pop:  pops[i],
		}
		moranEvents = append(moranEvents, event)
	}
	return
}

func generateEvents(popConfigs []pop.Config, pops []*pop.Pop, src rand.Source) (events []*pop.Event) {
	r := rand.New(src)
	for i := 0; i < len(popConfigs); i++ {
		c := popConfigs[i]

		mutateEvent := &pop.Event{
			Rate: c.Mutation.Rate * float64(c.Size*c.Length),
			Ops:  pop.NewSimpleMutator(r),
			Pop:  pops[i],
		}
		events = append(events, mutateEvent)

		inTransferEvent := &pop.Event{
			Rate: c.Transfer.In.Rate * float64(c.Size*c.Length),
			Ops:  pop.NewSimpleTransfer(c.Transfer.In.Fragment, r),
			Pop:  pops[i],
		}
		events = append(events, inTransferEvent)

		outTransferEvents := []*pop.Event{}
		totalSize := 0
		for j := 0; j < len(popConfigs); j++ {
			cj := popConfigs[j]
			if i != j {
				outE := &pop.Event{
					Rate: c.Transfer.Out.Rate * float64(c.Size*c.Length*cj.Size),
					Ops:  pop.NewOutTransfer(c.Transfer.Out.Fragment, pops[j], r),
					Pop:  pops[i],
				}
				totalSize += cj.Size
				outTransferEvents = append(outTransferEvents, outE)
			}
		}
		// to rescale the out transfer event rate.
		for j := 0; j < len(outTransferEvents); j++ {
			e := outTransferEvents[j]
			e.Rate = e.Rate / float64(totalSize) // cj.Size / totalSize
			events = append(events, e)
		}
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
