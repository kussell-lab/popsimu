package pop

type Sampler interface {
	Time(p *Pop) float64
	Operate(p *Pop)
	Wait()
	Start()
}
