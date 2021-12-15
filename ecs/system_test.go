package ecs

import (
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestSystem_Filter(t *testing.T) {
	w := NewWorld()

	s1 := &Component1System{}
	w.AddSystem(s1)

	t.Run("First nil check", func(t *testing.T) {
		w.Update(1)
		require.Len(t, s1.Filtered, 1)
		require.Len(t, s1.Filtered[0], 0)
	})

	// Add entities
	e1 := w.NewEntity()
	e1.Replace(&Component1{Num: 42})

	e2 := w.NewEntity()
	e2.Replace(&Component2{Text: "Hello world"})

	t.Run("After add entities 1 and 2", func(t *testing.T) {
		w.Update(1)
		require.Len(t, s1.Filtered, 1)
		require.Len(t, s1.Filtered[0], 1)
		require.Equal(t, e1.ID(), s1.Filtered[0][0].ID())
	})

	t.Run("Add exclude component to entity 1", func(t *testing.T) {
		e1.Replace(&Component2{Text: "Ops"})

		w.Update(1)
		require.Len(t, s1.Filtered, 1)
		require.Len(t, s1.Filtered[0], 0)
	})

	t.Run("After delete exclude component from entity 1", func(t *testing.T) {
		e1.Delete((*Component2)(nil))

		w.Update(1)
		require.Len(t, s1.Filtered, 1)

		f0 := s1.Filtered[0]
		require.Len(t, f0, 1)
		require.Equal(t, e1.ID(), f0[0].ID())
	})

	t.Run("After convert entity 2", func(t *testing.T) {
		e2.Delete((*Component2)(nil))
		e2.Replace(&Component1{Num: 1234})

		w.Update(1)
		require.Len(t, s1.Filtered, 1)

		f0 := s1.Filtered[0]
		require.Len(t, f0, 2)

		sort.Slice(f0, func(i, j int) bool {
			return f0[i].ID() < f0[j].ID()
		})

		require.Equal(t, e1.ID(), f0[0].ID())
		require.Equal(t, e2.ID(), f0[1].ID())
	})

	// Add new entity
	e3 := w.NewEntity()
	e3.Replace(&Component1{Num: 3333})

	t.Run("After add entity 3", func(t *testing.T) {
		w.Update(1)
		require.Len(t, s1.Filtered, 1)

		f0 := s1.Filtered[0]
		require.Len(t, f0, 3)

		sort.Slice(f0, func(i, j int) bool {
			return f0[i].ID() < f0[j].ID()
		})

		require.Equal(t, e1.ID(), f0[0].ID())
		require.Equal(t, e2.ID(), f0[1].ID())
		require.Equal(t, e3.ID(), f0[2].ID())
	})

	e3.Replace(&Component2{Text: "Hello world"})

	t.Run("After add Component2 to entity 3", func(t *testing.T) {
		w.Update(1)
		require.Len(t, s1.Filtered, 1)

		f0 := s1.Filtered[0]
		require.Len(t, f0, 2)

		sort.Slice(f0, func(i, j int) bool {
			return f0[i].ID() < f0[j].ID()
		})

		require.Equal(t, e1.ID(), f0[0].ID())
		require.Equal(t, e2.ID(), f0[1].ID())
	})

	s1n2 := &Component1And2System{}
	w.AddSystem(s1n2)

	t.Run("Add system 1And2", func(t *testing.T) {
		w.Update(1)

		require.Len(t, s1n2.Filtered, 1)

		f0 := s1n2.Filtered[0]
		require.Len(t, f0, 1)
		require.Equal(t, e3.ID(), f0[0].ID())
	})

	t.Run("Before delete systems", func(t *testing.T) {
		w.Update(1)

		require.Len(t, w.(*world).systems, 2)
		require.Len(t, w.(*world).systemFiltersEntityCache, 2)
		require.Len(t, w.(*world).systemFilters, 2)
	})

	w.RemoveSystem((*Component1System)(nil))
	w.RemoveSystem((*Component1And2System)(nil))

	t.Run("After delete systems", func(t *testing.T) {
		w.Update(1)

		require.Len(t, w.(*world).systems, 0)
		require.Len(t, w.(*world).systemFiltersEntityCache, 0)
		require.Len(t, w.(*world).systemFilters, 0)
	})

	s1or2 := &Component1Or2System{}
	w.AddSystem(s1or2)

	// revert entity 2
	e2.Delete((*Component1)(nil))
	e2.Replace(&Component2{Text: "Hello world"})

	t.Run("Add system 1Or2", func(t *testing.T) {
		w.Update(1)

		require.Len(t, s1or2.Filtered, 3)

		f0 := s1or2.Filtered[0]
		f1 := s1or2.Filtered[1]
		f2 := s1or2.Filtered[2]

		require.Len(t, f0, 1)
		require.Len(t, f1, 1)
		require.Len(t, f2, 1)

		require.Equal(t, e1.ID(), f0[0].ID())
		require.Equal(t, e2.ID(), f1[0].ID())
		require.Equal(t, e3.ID(), f2[0].ID())
	})
}

var _ System = (*Component1System)(nil)

type Component1System struct {
	Filtered [][]Entity
}

func (s *Component1System) GetFilters() []SystemFilter {
	return []SystemFilter{
		{[]Component{(*Component1)(nil)}, []Component{(*Component2)(nil)}},
	}
}

func (s *Component1System) Update(dt float32, filtered [][]Entity) {
	s.Filtered = filtered

	println("Component1System")
	for _, f := range filtered {
		for _, e := range f {
			spew.Dump(e.ID(), e.Get((*Component1)(nil)))
			spew.Dump(e.ID(), e.Get((*Component2)(nil)))
		}
	}
}

var _ System = (*Component2System)(nil)

type Component2System struct {
	Filtered [][]Entity
}

func (s *Component2System) GetFilters() []SystemFilter {
	return []SystemFilter{
		{[]Component{(*Component2)(nil)}, []Component{(*Component1)(nil)}},
	}
}

func (s *Component2System) Update(dt float32, filtered [][]Entity) {
	s.Filtered = filtered

	println("Component2System")
	for _, f := range filtered {
		for _, e := range f {
			spew.Dump(e.ID(), e.Get((*Component1)(nil)))
			spew.Dump(e.ID(), e.Get((*Component2)(nil)))
		}
	}
}

var _ System = (*Component1And2System)(nil)

type Component1And2System struct {
	Filtered [][]Entity
}

func (s *Component1And2System) GetFilters() []SystemFilter {
	return []SystemFilter{
		{[]Component{(*Component1)(nil), (*Component2)(nil)}, nil},
	}
}

func (s *Component1And2System) Update(dt float32, filtered [][]Entity) {
	s.Filtered = filtered

	println("Component1And2System")
	for _, f := range filtered {
		for _, e := range f {
			spew.Dump(e.ID(), e.Get((*Component1)(nil)))
			spew.Dump(e.ID(), e.Get((*Component2)(nil)))
		}
	}
}

var _ System = (*Component1Or2System)(nil)

type Component1Or2System struct {
	Filtered [][]Entity
}

func (s *Component1Or2System) GetFilters() []SystemFilter {
	return []SystemFilter{
		{[]Component{(*Component1)(nil)}, []Component{(*Component2)(nil)}},
		{[]Component{(*Component2)(nil)}, []Component{(*Component1)(nil)}},
		{[]Component{(*Component1)(nil), (*Component2)(nil)}, nil},
	}
}

func (s *Component1Or2System) Update(dt float32, filtered [][]Entity) {
	s.Filtered = filtered

	println("Component1Or2System")
	for _, f := range filtered {
		for _, e := range f {
			spew.Dump(e.ID(), e.Get((*Component1)(nil)))
			spew.Dump(e.ID(), e.Get((*Component2)(nil)))
		}
	}
}
