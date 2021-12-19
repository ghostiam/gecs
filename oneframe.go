package gecs

type oneFrame struct {
	c Component
}

// NewOneFrame returns a system that, when called, removes the component from all entities.
// Takes in a component whose type is to be removed.
//
// The use of this system involves adding it at the very end of the list of systems,
// so that the deletion occurs at the end of the cycle.
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
