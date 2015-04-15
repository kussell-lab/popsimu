package pop

import (
	"github.com/mingzhi/gomath/random"
	"github.com/mingzhi/gomath/stat/desc"
	"github.com/mingzhi/gsl-cgo/randist"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

type lockedSource struct {
	lk  sync.Mutex
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

func TestSingleMoran2(t *testing.T) {
	// set number of CPUs for using
	runtime.GOMAXPROCS(runtime.NumCPU())
	// population parameters
	popSizeArr := []int{10}
	mutRates := []float64{0.001, 0.01}
	traRates := []float64{0, 0.001, 0.01, 0.1}
	genomeLen := 100
	frag := 10
	replicates := 100

	// randome number source.
	src := &lockedSource{src: rand.NewSource(1)}
	r := rand.New(src)

	for i, popSize := range popSizeArr {
		mutRate := mutRates[i]
		numGens := 10 * popSize * popSize
		for _, traRate := range traRates {
			mean := desc.NewMean()
			vard := desc.NewVarianceWithBiasCorrection()
			for j := 0; j < replicates; j++ {
				p := New()
				p.Size = popSize
				p.Length = genomeLen
				p.Alphabet = []byte{1, 2, 3, 4}

				NewRandomPopGenerator(r).Operate(p)

				moranOps := NewMoranSampler(r)
				mutOps := NewSimpleMutator(mutRate, r)
				traOps := NewSimpleTransfer(traRate, frag, r)

				poisson := random.NewPoisson(float64(genomeLen)*(mutRate+traRate), src)

				p.OpsChan = make(chan Operator)
				go func() {
					defer close(p.OpsChan)
					for k := 0; k < numGens; k++ {
						p.OpsChan <- moranOps
						count := poisson.Int()
						for c := 0; c < count; c++ {
							v := r.Float64()
							if v <= mutRate/(mutRate+traRate) {
								p.OpsChan <- mutOps
							} else {
								p.OpsChan <- traOps
							}
						}
					}
				}()

				p.Evolve()

				d, _ := CalcKs(p)
				mean.Increment(d)
				vard.Increment(d)
			}

			res := mean.GetResult()
			ste := math.Sqrt(vard.GetResult() / float64(vard.GetN()))
			nu := float64(popSize) * mutRate
			gamma := float64(frag) * traRate
			exp := nu / (1 + gamma + 4.0/3.0*nu)
			if math.Abs(res-exp) > 2.0*ste {
				t.Errorf("n = %d, u = %f, t = %f, Expected %f, but got %f, at standard error %f\n", popSize, mutRate, traRate, exp, res, ste)
			}
		}
	}
}

func BenchmarkSingleMoran1(b *testing.B) {
	popSize := 100
	mutRate := 0.001
	traRate := 0.001
	genomeLen := 100
	frag := 10
	p := New()
	p.Size = popSize
	p.Length = genomeLen
	p.Alphabet = []byte{1, 2, 3, 4}

	// randome number source.
	src := &lockedSource{src: rand.NewSource(1)}
	r := rand.New(src)

	NewRandomPopGenerator(r).Operate(p)

	moranOps := NewMoranSampler(r)
	mutOps := NewSimpleMutator(mutRate, r)
	traOps := NewSimpleTransfer(traRate, frag, r)

	poisson := random.NewPoisson(float64(genomeLen)*(mutRate+traRate), src)

	p.OpsChan = make(chan Operator)
	go func() {
		defer close(p.OpsChan)
		for k := 0; k < b.N; k++ {
			p.OpsChan <- moranOps
			count := poisson.Int()
			for c := 0; c < count; c++ {
				v := r.Float64()
				if v <= mutRate/(mutRate+traRate) {
					p.OpsChan <- mutOps
				} else {
					p.OpsChan <- traOps
				}
			}
		}
	}()

	p.Evolve()
}

func BenchmarkSingleMoran2(b *testing.B) {
	popSize := 100
	mutRate := 0.001
	traRate := 0.001
	genomeLen := 100
	frag := 10
	p := New()
	p.Size = popSize
	p.Length = genomeLen
	p.Alphabet = []byte{1, 2, 3, 4}

	// randome number source.
	rng := randist.NewRNG(randist.MT19937)
	r := randist.NewUniform(rng)

	NewRandomPopGenerator(r).Operate(p)

	moranOps := NewMoranSampler(r)
	mutOps := NewSimpleMutator(mutRate, r)
	traOps := NewSimpleTransfer(traRate, frag, r)

	p.OpsChan = make(chan Operator)
	go func() {
		defer close(p.OpsChan)
		for k := 0; k < b.N; k++ {
			p.OpsChan <- moranOps
			count := randist.PoissonRandomInt(rng, float64(genomeLen)*(mutRate+traRate))
			for c := 0; c < count; c++ {
				v := r.Float64()
				if v <= mutRate/(mutRate+traRate) {
					p.OpsChan <- mutOps
				} else {
					p.OpsChan <- traOps
				}
			}
		}
	}()

	p.Evolve()
}

func TestSingleMoran(t *testing.T) {
	// set number of CPUs for using
	runtime.GOMAXPROCS(runtime.NumCPU())
	// population parameters
	popSizeArr := []int{10}
	mutRates := []float64{0.001, 0.01}
	traRates := []float64{0, 0.001, 0.01, 0.1}
	genomeLen := 100
	frag := 10
	replicates := 100

	// random number source
	rng := randist.NewRNG(randist.MT19937_1999)
	uniform := randist.NewUniform(rng)

	for i, popSize := range popSizeArr {
		mutRate := mutRates[i]
		numOfGens := 10 * popSize * popSize
		for _, traRate := range traRates {
			mean := desc.NewMean()
			vard := desc.NewVarianceWithBiasCorrection()
			for j := 0; j < replicates; j++ {
				p := New()
				p.Size = popSize
				p.Length = genomeLen
				p.Alphabet = []byte{1, 2, 3, 4}

				// operations
				popGenOps := NewRandomPopGenerator(uniform)
				moranOps := NewMoranSampler(uniform)
				mutOps := NewSimpleMutator(mutRate, uniform)
				traOps := NewSimpleTransfer(traRate, frag, uniform)

				operations := make(chan Operator)
				go func() {
					defer close(operations)
					// initialize the population
					popGenOps.Operate(p)
					for k := 0; k < numOfGens; k++ {
						operations <- moranOps
						tInterval := randist.ExponentialRandomFloat64(rng, 1.0/float64(popSize))
						totalRate := tInterval * float64(popSize*genomeLen) * (mutRate + traRate)
						count := randist.PoissonRandomInt(rng, totalRate)
						for c := 0; c < count; c++ {
							v := uniform.Float64()
							if v <= mutRate/(mutRate+traRate) {
								operations <- mutOps
							} else {
								operations <- traOps
							}
						}
					}
				}()

				Evolve(p, operations)
				d, _ := CalcKs(p)
				mean.Increment(d)
				vard.Increment(d)
			}

			res := mean.GetResult()
			ste := math.Sqrt(vard.GetResult() / float64(vard.GetN()))
			nu := float64(popSize) * mutRate
			gamma := float64(frag) * traRate
			exp := nu / (1 + gamma + 4.0/3.0*nu)
			if math.Abs(res-exp) > 2.0*ste {
				t.Errorf("n = %d, u = %f, t = %f, Expected %f, but got %f, at standard error %f\n", popSize, mutRate, traRate, exp, res, ste)
			}
		}
	}

}
