package corr

import (
	"math"
	"math/rand"
	"testing"
)

const tolerance float64 = 1e-5

func TestAutoCorr(t *testing.T) {
	data := []float64{
		0.1576, 0.9706, 0.9572, 0.4854, 0.8003, 0.1419, 0.4218, 0.9157, 0.7922, 0.9595,
	}

	expected := []float64{
		5.34401, 3.98031, 3.13718, 2.4438, 1.88223, 2.46069, 2.17929,
		1.83166, 1.05614, 0.151217,
	}

	res1 := AutoCorrBruteForce(data)
	res2 := AutoCorrFFT(data)
	if len(res1) != len(res2) {
		t.Errorf("Results 1 length of %d, results 2 length of %d\n", len(res1), len(res2))
	}
	for i := 0; i < len(expected); i++ {
		if math.Abs(res1[i]-res2[i]) > tolerance {
			t.Errorf("Result 1 %f, result 2 %f, at %d\n", res2[i], res1[i], i)
		}
		if math.Abs(res1[i]-expected[i]) > tolerance {
			t.Errorf("Expected %f, got %f, at %d\n", expected[i], res1[i], i)
		}
	}
}

func TestXCorr(t *testing.T) {
	data1 := []float64{
		0.6557,
		0.0357,
		0.8491,
		0.9340,
		0.6787,
		0.7577,
		0.7431,
		0.3922,
		0.6555,
		0.1712,
	}
	data2 := []float64{
		0.1576,
		0.9706,
		0.9572,
		0.4854,
		0.8003,
		0.1419,
		0.4218,
		0.9157,
		0.7922,
		0.9595,
	}
	expected := []float64{
		3.41092, 3.86624, 3.40214, 2.79604, 3.00792, 2.27675, 1.87809,
		1.44342, 0.5537, 0.629144,
	}

	res1 := XCorrFFT(data2, data1)
	res2 := XCorrBruteForce(data2, data1)
	if len(res1) != len(res2) {
		t.Errorf("Results 1 length of %d, results 2 length of %d\n", len(res1), len(res2))
	}
	for i := 0; i < len(expected); i++ {
		if math.Abs(res1[i]-res2[i]) > tolerance {
			t.Errorf("Result 1 %f, result 2 %f, at %d\n", res1[i], res2[i], i)
		}
		if math.Abs(res2[i]-expected[i]) > tolerance {
			t.Errorf("Expected %f, got %f, at %d\n", expected[i], res2[i], i)
		}
	}

}

func BenchmarkFFTAuto(b *testing.B) {
	data := make([]float64, 510)
	for i := 0; i < len(data); i++ {
		data[i] = rand.Float64()
	}

	for i := 0; i < b.N; i++ {
		AutoCorrFFT(data)
	}

}

func BenchmarkFFTBF(b *testing.B) {
	data := make([]float64, 510)
	for i := 0; i < len(data); i++ {
		data[i] = rand.Float64()
	}

	for i := 0; i < b.N; i++ {
		AutoCorrBruteForce(data)
	}

}
