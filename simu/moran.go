package simu

import (
	"math/rand"
	"time"

	"github.com/mingzhi/numgo/random"
	"github.com/mingzhi/popsimu/pop"
)

// Moran run simulations of multiple populations evolving under the Moran model.
func Moran(pops []*pop.Pop, popConfigs []pop.Config, numGen int) {
	randomSrc := rand.NewSource(time.Now().UnixNano())

	// Prepare a collection of possible events.
	events := generateEvents(popConfigs, pops, randomSrc)
	moranEvents := generateMoranEvents(popConfigs, pops, randomSrc)

	totalPopSize := 0
	for i := 0; i < len(pops); i++ {
		totalPopSize += pops[i].Size()
	}

	// total rate of all events scale by the total population size.
	totalRate := 0.0
	for i := 0; i < len(events); i++ {
		// the rate unit is per genome per generation,
		// so we need to rescale it by dividing the population size.
		totalRate += events[i].Rate / float64(totalPopSize)
	}

	r := random.New(randomSrc)
	rw := pop.NewRouletteWheel(randomSrc)
	eventChan := make(chan *pop.Event)
	go func() {
		defer close(eventChan)
		for i := 0; i < numGen; i++ {
			e := pop.Emit(moranEvents, rw)
			eventChan <- e
			eventCount := r.PoissonInt64(totalRate)
			var i int64
			for ; i < eventCount; i++ {
				e := pop.Emit(events, rw)
				eventChan <- e
			}
		}
	}()

	pop.Evolve(eventChan)
}

func generateEvents(popConfigs []pop.Config, pops []*pop.Pop, src rand.Source) (events []*pop.Event) {
	for i := 0; i < len(popConfigs); i++ {
		c := popConfigs[i]
		p := pops[i]

		mutateEvent := &pop.Event{
			Rate: c.Mutation.Rate * float64(p.Size()*c.Length),
			Ops:  pop.NewSimpleMutator([]byte(c.Alphabet), src),
			Pop:  pops[i],
		}
		events = append(events, mutateEvent)

		// choosing fragment size generator.
		var inFragGenerator pop.FragSizeGenerator
		switch c.FragGenerator {
		case "exponential":
			lambda := 1.0 / float64(c.Transfer.In.Fragment)
			inFragGenerator = pop.NewExpFrag(lambda, src)
		default:
			inFragGenerator = pop.NewConstantFrag(c.Transfer.In.Fragment)
		}

		inTransferEvent := &pop.Event{
			Rate: c.Transfer.In.Rate * float64(p.Size()*c.Length),
			Ops:  pop.NewSimpleTransfer(inFragGenerator, src),
			Pop:  pops[i],
		}
		events = append(events, inTransferEvent)

		// choosing fragment size generator.
		var outFragGenerator pop.FragSizeGenerator
		switch c.FragGenerator {
		case "exponential":
			lambda := 1.0 / float64(c.Transfer.Out.Fragment)
			inFragGenerator = pop.NewExpFrag(lambda, src)
		default:
			inFragGenerator = pop.NewConstantFrag(c.Transfer.In.Fragment)
		}

		outTransferEvents := []*pop.Event{}
		totalSize := 0
		for j := 0; j < len(popConfigs); j++ {
			pj := pops[j]
			if i != j {
				outE := &pop.Event{
					Rate: c.Transfer.Out.Rate * float64(p.Size()*c.Length*pj.Size()),
					Ops:  pop.NewOutTransfer(outFragGenerator, pj, src),
					Pop:  pops[i],
				}
				totalSize += pj.Size()
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
			Rate: float64(pops[i].Size()),
			Ops:  pop.NewMoranSampler(r),
			Pop:  pops[i],
		}
		moranEvents = append(moranEvents, event)
	}
	return
}
