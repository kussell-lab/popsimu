package pop

import (
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"time"
)

type sampleT func(p *Pop)

func CalcT2(p *Pop, sampleSize int) []float64 {
	t0 := time.Now()
	lineageChan := sampleLineages(p, sampleSize, 2)
	res := calcCoalTimes(lineageChan, p)
	t1 := time.Now()
	fmt.Printf("Calculate T2 using %v second.\n", t1.Sub(t0).Seconds())
	return res
}

func CalcT3(p *Pop, sampleSize int) []float64 {
	t0 := time.Now()
	lineageChan := sampleLineages(p, sampleSize, 3)
	res := calcCoalTimes(lineageChan, p)
	t1 := time.Now()
	fmt.Printf("Calculate T3 using %v second.\n", t1.Sub(t0).Seconds())
	return res
}

func CalcT4(p *Pop, sampleSize int) []float64 {
	t0 := time.Now()
	lineageChan := sampleLineages(p, sampleSize, 4)
	res := calcCoalTimes(lineageChan, p)
	t1 := time.Now()
	fmt.Printf("Calculate T4 using %v second.\n", t1.Sub(t0).Seconds())
	return res
}

func randomSample(list Lineages, n int) Lineages {
	m := make(map[int]bool)
	set := Lineages{}
	for i := len(list) - n; i < len(list); i++ {
		pos := rand.Intn(i + 1)
		if m[pos] {
			set = append(set, list[i])
			m[i] = true
		} else {
			set = append(set, list[pos])
			m[pos] = true
		}
	}
	return set
}

func sampleLineages(p *Pop, sampleSize, lineageSize int) chan Lineages {
	jobs := make(chan Lineages)
	go func() {
		defer close(jobs)
		for i := 0; i < sampleSize; i++ {
			ls := randomSample(p.Lineages, lineageSize)
			jobs <- ls
		}
	}()
	return jobs
}

func calcCoalTimes(c chan Lineages, p *Pop) []float64 {
	numWorker := runtime.GOMAXPROCS(0)
	results := make(chan float64, numWorker)
	done := make(chan bool)
	for i := 0; i < numWorker; i++ {
		go func() {
			for lineages := range c {
				var v float64
				t := findMostRecentCoalescentTime(lineages)
				v = float64(p.NumGeneration - t + 1)
				results <- v
			}
			done <- true
		}()
	}

	go func() {
		defer close(results)
		for i := 0; i < numWorker; i++ {
			<-done
		}
	}()

	coalTimes := []float64{}
	for v := range results {
		coalTimes = append(coalTimes, v)
	}

	return coalTimes
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
				a, b := lineages[i], lineages[j]
				if a.BirthTime == b.BirthTime && a.Parent == b.Parent {
					coalTimes = append(coalTimes, a.BirthTime)
				}
			}
		}

		if len(coalTimes) >= 1 {
			sort.Ints(coalTimes)
			return coalTimes[0]
		} else {
			sort.Sort(ByBirthTimeReverse{lineages})
			currentTime := lineages[0].BirthTime
			for i := 0; i < len(lineages); i++ {
				if lineages[i].BirthTime == currentTime {
					lineages[i] = lineages[i].Parent
				}
			}
			return findMostRecentCoalescentTime(lineages)
		}
	}
}
