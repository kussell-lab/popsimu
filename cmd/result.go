package cmd

import (
	"github.com/mingzhi/popsimu/pop"
)

type Result struct {
	Config pop.Config
	C      CovResult
}

type CovResult struct {
	Ks    float64
	KsN   int
	KsVar float64
	Ct    []float64
	CtN   []int
	CtVar []float64
}
