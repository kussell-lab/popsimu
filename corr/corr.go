package corr

import (
	"github.com/mjibson/go-dsp/fft"
	"math/cmplx"
)

func AutoCorrFFT(x []float64) []float64 {
	return XCorrFFT(x, x)
}

func XCorrFFT(x1, x2 []float64) []float64 {
	// zero padding.
	length := len(x1) * 2
	var ftlength int
	var i uint32 = 1
	for ftlength < length {
		ftlength = 1 << i
		i++
	}
	datax1 := make([]complex128, ftlength)
	datax2 := make([]complex128, ftlength)
	for i := 0; i < len(x1); i++ {
		datax1[i] = complex(x1[i], 0)
		datax2[i] = complex(x2[i], 0)
	}

	v1 := fft.FFT(datax1)
	v2 := fft.FFT(datax2)
	temp := []complex128{}
	for i := 0; i < len(v1); i++ {
		v := v1[i] * cmplx.Conj(v2[i])
		temp = append(temp, v)
	}
	v3 := fft.IFFT(temp)

	res := []float64{}
	for i := 0; i < len(x1); i++ {
		res = append(res, real(v3[i]))
	}
	return res
}

func XCorrBruteForce(x1, x2 []float64) []float64 {
	// zero padding.
	datax1 := make([]float64, len(x1)*2)
	datax2 := make([]float64, len(x2)*2)
	for i := 0; i < len(x1); i++ {
		datax1[i] = x1[i]
		datax2[i] = x2[i]
	}
	res := make([]float64, len(x1))
	for i := 0; i < len(x1); i++ {
		for k := 0; k < len(datax1); k++ {
			res[i] += datax1[k] * datax2[(len(datax2)+k-i)%len(datax2)]
		}
	}
	return res[:len(x1)]
}

func AutoCorrBruteForce(x []float64) []float64 {
	return XCorrBruteForce(x, x)
}
