package pop

import (
	"github.com/mingzhi/gomath/stat/desc"
)

type KsCalculator struct {
}

func (k *KsCalculator) Calc(p *Pop) (v float64) {
	m := desc.NewMean()
	for i := 0; i < p.Size; i++ {
		for j := i + 1; j < p.Size; j++ {
			for k := 0; k < p.Length; k++ {
				if p.Genomes[i][k] == p.Genomes[j][k] {
					m.Increment(0)
				} else {
					m.Increment(1)
				}
			}
		}
	}

	v = m.GetResult()

	return
}

func NewKsCalculator() *KsCalculator {
	return &KsCalculator{}
}
