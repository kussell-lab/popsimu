package pop

import (
	"math/rand"
	"runtime"

	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/gomath/stat/desc"
)

func CalcKs(sampleSize int, src rand.Source, p *Pop, others ...*Pop) (ks, vd float64) {
	events := []*Event{&Event{Rate: float64(p.Size()), Pop: p}}
	for i := 0; i < len(others); i++ {
		events = append(events, &Event{Rate: float64(others[i].Size()), Pop: others[i]})
	}

	rw := NewRouletteWheel(src)
	r := rand.New(src)

	matrix := [][]float64{}
	for s := 0; s < sampleSize; s++ {
		p1 := Emit(events, rw).Pop
		p2 := Emit(events, rw).Pop
		i, j := r.Intn(p1.Size()), r.Intn(p2.Size())
		x := make([]float64, p.Length())
		for k := 0; k < p.Length(); k++ {
			if p1.Genomes[i].Seq()[k] != p2.Genomes[j].Seq()[k] {
				x[k] = 1
			}
		}
		matrix = append(matrix, x)
	}
	ks, vd = calcKs(matrix)
	return
}

func calcKs(matrix [][]float64) (ks, vd float64) {
	m := desc.NewMean()
	v := desc.NewVariance()

	for i := 0; i < len(matrix); i++ {
		mean := desc.NewMean()
		for j := 0; j < len(matrix[i]); j++ {
			mean.Increment(matrix[i][j])
		}
		m.Increment(mean.GetResult())
		v.Increment(mean.GetResult())
	}
	ks = m.GetResult()
	vd = v.GetResult()
	return
}

func CrossKs(sampleSize int, src rand.Source, p1, p2 *Pop) (ks, vd float64) {
	r := rand.New(src)
	matrix := [][]float64{}
	for s := 0; s < sampleSize; s++ {
		i, j := r.Intn(p1.Size()), r.Intn(p2.Size())
		x := make([]float64, p1.Length())
		for k := 0; k < p1.Length(); k++ {
			if p1.Genomes[i].Seq()[k] != p2.Genomes[j].Seq()[k] {
				x[k] = 1
			}
		}
		matrix = append(matrix, x)
	}
	ks, vd = calcKs(matrix)

	return
}

func CalcCov(sampleSize, maxL int, src rand.Source, p *Pop, others ...*Pop) (cm, ct, cr, cs []float64) {
	matrix := [][]float64{}
	if maxL > p.Length() {
		maxL = p.Length()
	}

	events := []*Event{&Event{Rate: float64(p.Size()), Pop: p}}
	for i := 0; i < len(others); i++ {
		events = append(events, &Event{Rate: float64(others[i].Size()), Pop: others[i]})
	}

	rw := NewRouletteWheel(src)
	r := rand.New(src)

	for s := 0; s < sampleSize; s++ {
		p1 := Emit(events, rw).Pop
		p2 := Emit(events, rw).Pop
		i, j := r.Intn(p1.Size()), r.Intn(p2.Size())
		profile := []float64{}
		for k := 0; k < p.Length(); k++ {
			if p1.Genomes[i].Seq()[k] != p2.Genomes[j].Seq()[k] {
				profile = append(profile, 1)
			} else {
				profile = append(profile, 0)
			}
		}
		matrix = append(matrix, profile)
	}

	circular := true
	return calcCov(matrix, p.Length(), circular)
}

func CrossCov(sampleSize, maxL int, src rand.Source, p1, p2 *Pop) (cm, ct, cr, cs []float64) {
	matrix := [][]float64{}
	if maxL > p1.Length() {
		maxL = p1.Length()
	}

	r := rand.New(src)
	for s := 0; s < sampleSize; s++ {
		i := r.Intn(p1.Size())
		j := r.Intn(p2.Size())
		profile := []float64{}
		for k := 0; k < p1.Length(); k++ {
			if p1.Genomes[i].Seq()[k] != p2.Genomes[j].Seq()[k] {
				profile = append(profile, 1)
			} else {
				profile = append(profile, 0)
			}
		}
		matrix = append(matrix, profile)
	}

	circular := true

	return calcCov(matrix, p1.Length(), circular)
}

func calcCov(matrix [][]float64, maxL int, circular bool) (cm, ct, cr, cs []float64) {
	cm, ct = calcCmFFT(matrix, maxL, circular)
	cr, _ = calcCmFFT([][]float64{average(matrix)}, maxL, circular)
	for i := 0; i < len(ct); i++ {
		cs = append(cs, ct[i]-cr[i])
	}
	return
}

func calcCmFFT(matrix [][]float64, maxL int, circular bool) (mutCov, totCov []float64) {
	// mask
	mask := make([]float64, len(matrix[0]))
	for i := 0; i < len(mask); i++ {
		mask[i] = 1.0
	}
	maskCorr := correlation.AutoCorrFFT(mask, circular)

	jobs := make(chan []float64)
	go func() {
		defer close(jobs)
		for i := 0; i < len(matrix); i++ {
			jobs <- matrix[i]
		}
	}()

	type result struct {
		mean *desc.Mean
		pxy  []float64
	}
	resChan := make(chan result)
	done := make(chan bool)
	ncpu := runtime.GOMAXPROCS(0)
	for i := 0; i < ncpu; i++ {
		go func() {
			for x := range jobs {
				xy := correlation.AutoCorrFFT(x, circular)
				pxy := make([]float64, len(xy))
				for i := 0; i < len(xy); i++ {
					pxy[i] = (xy[i] + xy[(len(xy)-i)%len(xy)]) / (maskCorr[i] + maskCorr[(len(xy)-i)%len(maskCorr)])
				}

				mean := desc.NewMean()
				for i := 0; i < len(x); i++ {
					mean.Increment(x[i])
				}

				resChan <- result{mean: mean, pxy: pxy}
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

	means1 := make([]*desc.Mean, maxL)
	means2 := make([]*desc.Mean, maxL)
	for i := 0; i < maxL; i++ {
		means1[i] = desc.NewMean()
		means2[i] = desc.NewMean()
	}

	totMean := desc.NewMean()
	for res := range resChan {
		mean := res.mean
		pxy := res.pxy
		meanSquare := mean.GetResult() * mean.GetResult()
		totMean.Append(mean)
		for j := 0; j < maxL; j++ {
			means1[j].Increment(pxy[j] - meanSquare)
			means2[j].Increment(pxy[j])
		}
	}

	totalSquare := totMean.GetResult() * totMean.GetResult()

	for i := 0; i < maxL; i++ {
		mutCov = append(mutCov, means1[i].GetResult())
		totCov = append(totCov, means2[i].GetResult()-totalSquare)
	}

	return
}

func calcCm(matrix [][]float64, maxL int, circular bool) (mutCov, totCov []float64) {
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
					for j := 0; j < len(matrix[i]); j++ {
						x, y := matrix[i][j], matrix[i][(j+l)%len(matrix[i])]
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

func calcCs(matrix [][]float64, maxL int, circular bool) []float64 {
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
				length := len(matrix[0]) - l
				if circular {
					length = len(matrix[0])
				}
				for j := 0; j < length; j++ {
					estimator := correlation.NewBivariateCovariance(biasCorrected)
					for i := 0; i < len(matrix); i++ {
						x, y := matrix[i][j], matrix[i][(j+l)%len(matrix[0])]
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
