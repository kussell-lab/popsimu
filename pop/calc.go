package pop

import (
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
	vard = m.GetResult()

	return
}
