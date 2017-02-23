package pop

import (
	"math/rand"
)

// Shock is an interface for challenging population with shocks.
type Shock interface {
	Reduce(p *Pop) *Pop
}

// Dilution reduces the population by certain poportion.
type Dilution struct {
	Factor float64
}

// Reduce reduces the population to certain poportion.
func (d Dilution) Reduce(p *Pop) *Pop {
	finalSize := int(float64(p.Size()) * d.Factor)
	indices := make([]int, p.Size())
	for i := 0; i < p.Size(); i++ {
		indices[i] = i
	}
	shuffle(indices)
	var finalGenomes []Genome
	var finalLineages []*Lineage
	for i := 0; i < finalSize; i++ {
		finalGenomes = append(finalGenomes, p.Genomes[indices[i]])
		finalLineages = append(finalLineages, p.Lineages[indices[i]])
	}

	finalP := Pop{}
	finalP.Circled = p.Circled
	finalP.Genomes = finalGenomes
	finalP.Lineages = finalLineages
	finalP.NumGeneration = 0
	finalP.TargetSize = p.TargetSize

	return &finalP
}

// shuffle an int array
func shuffle(a []int) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

// Recover recovers the population exponentially.
func Recover(p *Pop, finalSize int) *Pop {
	for p.Size() < finalSize {
		index := rand.Intn(p.Size())
		genome := p.Genomes[index]
		daughter := genome.Copy()
		p.Genomes = append(p.Genomes, daughter)
		p.Lineages = append(p.Lineages, nil)
		p.Lineages[index], p.Lineages[p.Size()-1] = createNewLineages(p.Lineages[index], p.NumGeneration)
	}
	return p
}
