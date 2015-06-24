package pop

// Objects implementing the Operator interface
// can be used to do operations to a population.
type Operator interface {
	Operate(*Pop)
}
