package pop

import (
	"math"
	"math/rand"
	"runtime"
	"testing"
)

func TestCalcCm(t *testing.T) {
	tolerance := 1e-8
	size := 10
	length := 64
	matrix := make([][]float64, size)
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < length; j++ {
			matrix[i] = append(matrix[i], rand.Float64())
		}
	}
	circular := true
	mutCov1, totCov1 := calcCm(matrix, len(matrix[0]), circular)
	mutCov2, totCov2 := calcCmFFT(matrix, len(matrix[0]), circular)

	if len(mutCov1) != len(mutCov2) {
		t.Errorf("Different length of results (mutCov): %d vs %d\n", len(mutCov1), len(mutCov2))
	}

	if len(totCov1) != len(totCov2) {
		t.Errorf("Different length of results (totCov): %d vs %d\n", len(totCov1), len(totCov2))
	}

	for i := 0; i < len(matrix[0]); i++ {
		if math.IsNaN(mutCov1[i]) {
			t.Errorf("NaN at %d\n", i)
		}
		if math.IsNaN(mutCov2[i]) {
			t.Errorf("NaN at %d\n", i)
		}
		if math.IsNaN(totCov1[i]) {
			t.Errorf("NaN at %d\n", i)
		}
		if math.IsNaN(totCov2[i]) {
			t.Errorf("NaN at %d\n", i)
		}
		if math.Abs(mutCov1[i]-mutCov2[i]) > tolerance {
			t.Errorf("Difference between result1 %f and result2 %f at %d\n", mutCov1[i], mutCov2[i], i)
		}
		if math.Abs(totCov1[i]-totCov2[i]) > tolerance {
			t.Errorf("Difference between result1 %f and result2 %f at %d\n", totCov1[i], totCov2[i], i)
		}
	}
}

func TestDecomposition(t *testing.T) {
	tolerance := 1e-8
	size := 10
	length := 64
	matrix := make([][]float64, size)
	for i := 0; i < size; i++ {
		for j := 0; j < length; j++ {
			matrix[i] = append(matrix[i], rand.Float64())
		}
	}
	circular := true
	maxL := len(matrix[0])
	cM, cT := calcCmFFT(matrix, maxL, circular)
	cR, _ := calcCmFFT([][]float64{average(matrix)}, maxL, circular)
	cS := calcCs(matrix, maxL, circular)
	_, vd := calcKs(matrix)
	for i := 0; i < len(cM); i++ {
		vd1 := cT[i] - cM[i]
		if math.Abs(vd1-vd) > tolerance {
			t.Errorf("Difference between vd %f and vd1 %f at %d\n", vd, vd1, i)
		}
		cs1 := cS[i]
		cs2 := cT[i] - cR[i]
		if math.Abs(cs1-cs2) > tolerance {
			t.Errorf("Difference between cS %f and (cT - cR) %f at %d\n", cs1, cs2, i)
		}
	}
}

func benchmarkCalcCmFFT(length int, b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	size := 10
	matrix := make([][]float64, size)
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < length; j++ {
			matrix[i] = append(matrix[i], rand.Float64())
		}
	}
	circular := true
	for i := 0; i < b.N; i++ {
		calcCmFFT(matrix, length, circular)
	}
}

func BenchmarkCalcCmFFT1048(b *testing.B) {
	benchmarkCalcCmFFT(1048, b)
}

func BenchmarkCalcCmFFT1000(b *testing.B) {
	benchmarkCalcCmFFT(1000, b)
}

func BenchmarkCalcCm(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	size := 10
	length := 1048
	matrix := make([][]float64, size)
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < length; j++ {
			matrix[i] = append(matrix[i], rand.Float64())
		}
	}
	circular := true
	for i := 0; i < b.N; i++ {
		calcCm(matrix, length, circular)
	}
}

func BenchmarkCalcCs(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	size := 10
	length := 1048
	matrix := make([][]float64, size)
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < length; j++ {
			matrix[i] = append(matrix[i], rand.Float64())
		}
	}
	circular := true
	for i := 0; i < b.N; i++ {
		calcCs(matrix, length, circular)
	}
}
