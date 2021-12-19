package gecs

type oneFrame struct {
	c Component
}

func NewOneFrame(c Component) System {
	return &oneFrame{
		c: c,
	}
}

func (s *oneFrame) GetFilters() []SystemFilter {
	return []SystemFilter{{Include: []Component{s.c}}}
}

func (s *oneFrame) Update(_ float32, filtered [][]Entity) {
	for _, es := range filtered {
		for _, e := range es {
			e.Delete(s.c)
		}
	}
}
