package pop

import (
	"github.com/mingzhi/gomath/stat/desc/meanvar"
	"runtime"
	"sort"
)

func CalcT2(p *Pop) (mv *meanvar.MeanVar) {
	jobs := make(chan Lineages)
	go func() {
		for i := 0; i < len(p.Lineages); i++ {
			for j := i + 1; j < len(p.Lineages); j++ {
				lineages := Lineages{p.Lineages[i], p.Lineages[j]}
				jobs <- lineages
			}
		}
	}()

	return calcCoalTime(jobs, p)
}

func CalcT3(p *Pop) (mv *meanvar.MeanVar) {
	jobs := make(chan Lineages)
	go func() {
		defer close(jobs)
		for i := 0; i < len(p.Lineages); i++ {
			for j := i + 1; j < len(p.Lineages); j++ {
				for k := j + 1; k < len(p.Lineages); k++ {
					lineages := Lineages{p.Lineages[i], p.Lineages[j], p.Lineages[k]}
					jobs <- lineages
				}
			}
		}
	}()

	return calcCoalTime(jobs, p)
}

func CalcT4(p *Pop) (mv *meanvar.MeanVar) {
	jobs := make(chan Lineages)
	go func() {
		defer close(jobs)
		for i := 0; i < len(p.Lineages); i++ {
			for j := i + 1; j < len(p.Lineages); j++ {
				for k := j + 1; k < len(p.Lineages); k++ {
					for h := k + 1; h < len(p.Lineages); h++ {
						lineages := Lineages{p.Lineages[i], p.Lineages[j], p.Lineages[k], p.Lineages[h]}
						jobs <- lineages
					}
				}
			}
		}
	}()

	return calcCoalTime(jobs, p)
}

func calcCoalTime(c chan Lineages, p *Pop) (mv *meanvar.MeanVar) {
	numWorker := runtime.GOMAXPROCS(0)
	results := make(chan *meanvar.MeanVar, numWorker)
	for i := 0; i < numWorker; i++ {
		go func() {
			mv := meanvar.New()
			for lineages := range c {
				t := findMostRecentCoalescentTime(lineages)
				v := float64(p.NumGeneration - t + 1)
				mv.Increment(v)
			}
			results <- mv
		}()
	}

	for i := 0; i < numWorker; i++ {
		m := <-results
		if i == 0 {
			mv = m
		} else {
			mv.Append(m)
		}
	}

	return
}

func findMostRecentCoalescentTime(lineages Lineages) int {
	if len(lineages) == 0 {
		return 0
	} else if len(lineages) == 1 {
		return lineages[0].BirthTime
	} else {
		coalTimes := []int{}
		for i := 0; i < len(lineages); i++ {
			for j := i + 1; j < len(lineages); j++ {
				if lineages[i].BirthTime == lineages[j].BirthTime {
					coalTimes = append(coalTimes, lineages[i].BirthTime)
				}
			}
		}

		if len(coalTimes) >= 1 {
			sort.Ints(coalTimes)
			return coalTimes[0]
		} else {
			sort.Sort(ByBirthTimeReverse{lineages})
			lineages[0] = lineages[0].Parent
			return findMostRecentCoalescentTime(lineages)
		}
	}
}
