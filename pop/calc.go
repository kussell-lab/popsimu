package pop

import (
	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/gomath/stat/desc"
)

func CalcKs(p *Pop) (ks, vard float64) {
	m := desc.NewMean()
	v := desc.NewVarianceWithBiasCorrection()
	for i := 0; i < p.Size; i++ {
		for j := i + 1; j < p.Size; j++ {
			m1 := desc.NewMean() // average distance between two sequences.
			for k := 0; k < p.Length; k++ {
				if p.Genomes[i].Sequence[k] == p.Genomes[j].Sequence[k] {
					m1.Increment(0)
				} else {
					m1.Increment(1)
				}
			}
			m.Increment(m1.GetResult())
			v.Increment(m1.GetResult())
		}
	}

	ks = m.GetResult()
	vard = v.GetResult()

	return
}

func CrossKs(p1, p2 *Pop) (ks, vard float64) {
	m := desc.NewMean()
	v := desc.NewVarianceWithBiasCorrection()
	for i := 0; i < p1.Size; i++ {
		for j := 0; j < p2.Size; j++ {
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
	}

	ks = m.GetResult()
	vard = v.GetResult()

	return
}

func CalcCov(p *Pop) (cm, ct, cr, cs []float64) {
	matrix := [][]float64{}
	for i := 0; i < p.Size; i++ {
		for j := i + 1; j < p.Size; j++ {
			profile := []float64{}
			for k := 0; k < p.Length; k++ {
				if p.Genomes[i].Sequence[k] != p.Genomes[j].Sequence[k] {
					profile = append(profile, 1)
				} else {
					profile = append(profile, 0)
				}
			}
			matrix = append(matrix, profile)
		}
	}

	return calcCov(matrix, p.Length)
}

func CrossCov(p1, p2 *Pop) (cm, ct, cr, cs []float64) {
	matrix := [][]float64{}
	for i := 0; i < p1.Size; i++ {
		for j := 0; j < p2.Size; j++ {
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
	cms := []float64{}
	cts := []float64{}
	biasCorrected := false
	for l := 0; l < maxL; l++ {
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
		cms = append(cms, mean.GetResult())
		cts = append(cts, cov.GetResult())
	}

	mutCov = cms
	totCov = cts

	return
}

func calcCS(matrix [][]float64, maxL int) []float64 {
	css := []float64{}
	biasCorrected := false
	for l := 0; l < maxL; l++ {
		mean := desc.NewMean()
		for j := 0; j < len(matrix[0])-l; j++ {
			estimator := correlation.NewBivariateCovariance(biasCorrected)
			for i := 0; i < len(matrix); i++ {
				x, y := matrix[i][j], matrix[i][j+l]
				estimator.Increment(x, y)
			}
			mean.Increment(estimator.GetResult())
		}
		css = append(css, mean.GetResult())
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
