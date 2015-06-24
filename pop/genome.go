package pop

type Genome interface {
	Seq() []byte
	Fitness() float64
	Length() int
	Copy() Genome
}
