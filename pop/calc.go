package pop

import (
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/gomath/stat/desc"
	"math/rand"
	"runtime"
)

func CalcKs(sampleSize int, p *Pop, others ...*Pop) (ks, vard float64) {
	m := desc.NewMean()
	v := desc.NewVarianceWithBiasCorrection()

	events := []*Event{&Event{Rate: float64(p.Size), Pop: p}}
	for i := 0; i < len(others); i++ {
		events = append(events, &Event{Rate: float64(others[i].Size), Pop: others[i]})
	}
	r := rand.New(random.NewLockedSource(rand.NewSource(1)))

	for s := 0; s < sampleSize; s++ {
		p1 := Emit(events, r).Pop
		p2 := Emit(events, r).Pop
		i, j := rand.Intn(p1.Size), rand.Intn(p2.Size)
		m1 := desc.NewMean() // average distance between two sequences.
		for k := 0; k < p.Length; k++ {
			if p1.Genomes[i].Sequence[k] == p2.Genomes[j].Sequence[k] {
				m1.Increment(0)
			} else {
				m1.Increment(1)
			}
		}
		m.Increment(m1.GetResult())
		v.Increment(m1.GetResult())
	}

	ks = m.GetResult()
	vard = v.GetResult()

	return
}

func CrossKs(sampleSize int, p1, p2 *Pop) (ks, vard float64) {
	m := desc.NewMean()
	v := desc.NewVarianceWithBiasCorrection()
	for s := 0; s < sampleSize; s++ {
		i, j := rand.Intn(p1.Size), rand.Intn(p2.Size)
		m1 := desc.NewMean() // average distance between two sequences.
		for k := 0; k < p1.Length; k++ {
			if p1.Genomes[i].Sequence[k] == p2.Genomes[j].Sequence[k] {
				m1.Increment(0)
			} else {
				m1.Increment(1)
			}
		}
		m.Increment(m1.GetResult())
		v.Increment(m1.GetResult())
	}

	ks = m.GetResult()
	vard = v.GetResult()

	return
}

func CalcCov(sampleSize, maxL int, p *Pop, others ...*Pop) (cm, ct, cr, cs []float64) {
	matrix := [][]float64{}
	if maxL > p.Length {
		maxL = p.Length
	}

	events := []*Event{&Event{Rate: float64(p.Size), Pop: p}}
	for i := 0; i < len(others); i++ {
		events = append(events, &Event{Rate: float64(others[i].Size), Pop: others[i]})
	}
	r := rand.New(random.NewLockedSource(rand.NewSource(1)))

	for s := 0; s < sampleSize; s++ {
		p1 := Emit(events, r).Pop
		p2 := Emit(events, r).Pop
		i, j := rand.Intn(p1.Size), rand.Intn(p2.Size)
		profile := []float64{}
		for k := 0; k < p.Length; k++ {
			if p1.Genomes[i].Sequence[k] != p2.Genomes[j].Sequence[k] {
				profile = append(profile, 1)
			} else {
				profile = append(profile, 0)
			}
		}
		matrix = append(matrix, profile)
	}

	return calcCov(matrix, p.Length)
}

func CrossCov(sampleSize, maxL int, p1, p2 *Pop) (cm, ct, cr, cs []float64) {
	matrix := [][]float64{}
	if maxL > p1.Length {
		maxL = p1.Length
	}
	for s := 0; s < sampleSize; s++ {
		i := rand.Intn(p1.Size)
		j := rand.Intn(p2.Size)
		profile := []float64{}
		for k := 0; k < p1.Length; k++ {
			if p1.Genomes[i].Sequence[k] != p2.Genomes[j].Sequence[k] {
				profile = append(profile, 1)
			} else {
				profile = append(profile, 0)
			}
		}
		matrix = append(matrix, profile)
	}

	return calcCov(matrix, p1.Length)
}

func calcCov(matrix [][]float64, maxL int) (cm, ct, cr, cs []float64) {
	cm, ct = calcCM(matrix, maxL)
	cs = calcCS(matrix, maxL)
	cr, _ = calcCM([][]float64{average(matrix)}, maxL)
	return
}

func calcCM(matrix [][]float64, maxL int) (mutCov, totCov []float64) {
	cms := make([]float64, maxL)
	cts := make([]float64, maxL)
	biasCorrected := false

	ncpu := runtime.GOMAXPROCS(0)

	jobs := make(chan int)
	go func() {
		defer close(jobs)
		for i := 0; i < maxL; i++ {
			jobs <- i
		}
	}()

	type result struct {
		l    int
		mean *desc.Mean
		cov  *correlation.BivariateCovariance
	}

	done := make(chan bool)
	resChan := make(chan result)
	for c := 0; c < ncpu; c++ {
		go func() {
			for l := range jobs {
				mean := desc.NewMean()
				cov := correlation.NewBivariateCovariance(biasCorrected)
				for i := 0; i < len(matrix); i++ {
					estimator := correlation.NewBivariateCovariance(biasCorrected)
					for j := 0; j < len(matrix[i])-l; j++ {
						x, y := matrix[i][j], matrix[i][j+l]
						estimator.Increment(x, y)
					}
					mean.Increment(estimator.GetResult())
					cov.Append(estimator)
				}
				resChan <- result{l: l, mean: mean, cov: cov}
			}
			done <- true
		}()
	}

	go func() {
		defer close(resChan)
		for i := 0; i < ncpu; i++ {
			<-done
		}
	}()

	for res := range resChan {
		cms[res.l] = res.mean.GetResult()
		cts[res.l] = res.cov.GetResult()
	}

	mutCov = cms
	totCov = cts

	return
}

func calcCS(matrix [][]float64, maxL int) []float64 {
	css := make([]float64, maxL)
	biasCorrected := false

	jobs := make(chan int)
	go func() {
		defer close(jobs)
		for i := 0; i < maxL; i++ {
			jobs <- i
		}
	}()

	type result struct {
		l    int
		mean *desc.Mean
	}

	resChan := make(chan result)
	done := make(chan bool)

	ncpu := runtime.GOMAXPROCS(0)
	for i := 0; i < ncpu; i++ {
		go func() {
			for l := range jobs {
				mean := desc.NewMean()
				for j := 0; j < len(matrix[0])-l; j++ {
					estimator := correlation.NewBivariateCovariance(biasCorrected)
					for i := 0; i < len(matrix); i++ {
						x, y := matrix[i][j], matrix[i][j+l]
						estimator.Increment(x, y)
					}
					mean.Increment(estimator.GetResult())
				}
				resChan <- result{l: l, mean: mean}
			}
			done <- true
		}()
	}

	go func() {
		defer close(resChan)
		for i := 0; i < ncpu; i++ {
			<-done
		}
	}()

	for res := range resChan {
		css[res.l] = res.mean.GetResult()
	}

	return css
}

func average(matrix [][]float64) []float64 {
	profile := []float64{}
	for l := 0; l < len(matrix[0]); l++ {
		mean := desc.NewMean()
		for i := 0; i < len(matrix); i++ {
			mean.Increment(matrix[i][l])
		}
		profile = append(profile, mean.GetResult())
	}
	return profile
}
