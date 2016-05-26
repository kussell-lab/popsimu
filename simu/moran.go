package simu

import (
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/popsimu/pop"
	"math/rand"
	"time"
)

func RunMoran(pops []*pop.Pop, popConfigs []pop.Config, numGen int) {
	// Random source.
	seed := time.Now().UnixNano()
	src := random.NewLockedSource(rand.NewSource(seed))

	// Prepare possible events.
	events := generateEvents(popConfigs, pops, src)
	moranEvents := generateMoranEvents(popConfigs, pops, src)

	totalSize := 0
	for i := 0; i < len(pops); i++ {
		totalSize += pops[i].Size
	}

	totalRate := 0.0
	for i := 0; i < len(events); i++ {
		// the rate unit is per genome per generation,
		// so we need to rescale it by dividing the population size.
		totalRate += events[i].Rate / float64(totalSize)
	}

	// Prepare Poisson random source.
	poisson := random.NewPoisson(totalRate, src)
	r := rand.New(src)

	// Create an event channel and load it.
	eventChan := make(chan *pop.Event)
	go func() {
		defer close(eventChan)
		for i := 0; i < numGen; i++ {
			e := pop.Emit(moranEvents, r)
			eventChan <- e
			eventCount := poisson.Int()
			for i := 0; i < eventCount; i++ {
				e := pop.Emit(events, r)
				eventChan <- e
			}
		}
	}()

	pop.Evolve(eventChan)

	return pops
}

func generateEvents(popConfigs []pop.Config, pops []*pop.Pop, src rand.Source) (events []*pop.Event) {
	r := rand.New(src)
	for i := 0; i < len(popConfigs); i++ {
		c := popConfigs[i]
		p := pops[i]

		mutateEvent := &pop.Event{
			Rate: c.Mutation.Rate * float64(p.Size*c.Length),
			Ops:  pop.NewSimpleMutator(r),
			Pop:  pops[i],
		}
		events = append(events, mutateEvent)

		inTransferEvent := &pop.Event{
			Rate: c.Transfer.In.Rate * float64(p.Size*c.Length),
			Ops:  pop.NewSimpleTransfer(c.Transfer.In.Fragment, r),
			Pop:  pops[i],
		}
		events = append(events, inTransferEvent)

		outTransferEvents := []*pop.Event{}
		totalSize := 0
		for j := 0; j < len(popConfigs); j++ {
			pj := pops[j]
			if i != j {
				outE := &pop.Event{
					Rate: c.Transfer.Out.Rate * float64(p.Size*c.Length*pj.Size),
					Ops:  pop.NewOutTransfer(c.Transfer.Out.Fragment, pops[j], r),
					Pop:  pops[i],
				}
				totalSize += pj.Size
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

func generateMoranEvents(popConfigs []pop.Config, pops []*pop.Pop, src rand.Source) (moranEvents []*pop.Event) {
	r := rand.New(src)
	for i := 0; i < len(popConfigs); i++ {
		event := &pop.Event{
			Rate: float64(pops[i].Size),
			Ops:  pop.NewMoranSampler(r),
			Pop:  pops[i],
		}
		moranEvents = append(moranEvents, event)
	}
	return
}
