package main

import (
	"github.com/mingzhi/gomath/stat/desc"
)

type MeanVar struct {
	Mean *desc.Mean
	Var  *desc.Variance
}

func (m *MeanVar) Increment(d float64) {
	m.Mean.Increment(d)
	m.Var.Increment(d)
}

func NewMeanVar() *MeanVar {
	mv := MeanVar{}
	mv.Mean = desc.NewMean()
	mv.Var = desc.NewVarianceWithBiasCorrection()

	return &mv
}
