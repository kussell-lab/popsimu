package pop

type Rand interface {
	Intn(n int) int
	Float64() float64
}
