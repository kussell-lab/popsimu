package pop

type Lineage struct {
	BirthTime int // when it was produced.
	Parent    *Lineage
}

type Lineages []*Lineage

func (s Lineages) Len() int      { return len(s) }
func (s Lineages) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByBirthTimeReverse struct{ Lineages }

func (s ByBirthTimeReverse) Less(i, j int) bool {
	return s.Lineages[i].BirthTime > s.Lineages[j].BirthTime
}
