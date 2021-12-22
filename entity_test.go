package gecs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Component1 struct {
	Num int
}

type Component2 struct {
	Text string
}

func TestEntity_Get(t *testing.T) {
	w := NewWorld()

	t.Run("Get nil component", func(t *testing.T) {
		e := w.NewEntity()
		got := e.Get(nil)
		require.Equal(t, nil, got)
	})

	t.Run("Get nil type component", func(t *testing.T) {
		e := w.NewEntity()
		got := e.Get((*Component1)(nil))
		require.Equal(t, nil, got)
	})

	t.Run("Create not exist component", func(t *testing.T) {
		const expected = 42

		e := w.NewEntity()
		got := e.Get(&Component1{Num: 42}).(*Component1)
		require.Equal(t, expected, got.Num)

		t.Run("Get exist component", func(t *testing.T) {
			got = e.Get(&Component1{Num: 1234}).(*Component1)
			require.Equal(t, expected, got.Num, "Should return the value from the first added component")
		})

		t.Run("Get exist component by nil type", func(t *testing.T) {
			got = e.Get((*Component1)(nil)).(*Component1)
			require.Equal(t, expected, got.Num, "Should return the value from the first added component")
		})
	})
}

func TestEntity_Has(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	e.Get(&Component1{Num: 42})

	t.Run("Not has nil", func(t *testing.T) {
		require.False(t, e.Has(nil))
	})

	t.Run("Has by Component1 nil type", func(t *testing.T) {
		require.True(t, e.Has((*Component1)(nil)))
	})

	t.Run("Has by Component1 type", func(t *testing.T) {
		require.True(t, e.Has(&Component1{}))
	})

	t.Run("Not has by Component2 nil type", func(t *testing.T) {
		require.False(t, e.Has((*Component2)(nil)))
	})

	t.Run("Not has by Component2 type", func(t *testing.T) {
		require.False(t, e.Has(&Component2{}))
	})
}

func TestEntity_Replace(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	e.Get(&Component1{Num: 42})

	t.Run("Replace nil", func(t *testing.T) {
		e.Replace(nil)
	})

	t.Run("Replace nil type", func(t *testing.T) {
		e.Replace((*Component1)(nil))
	})

	t.Run("Replace 1", func(t *testing.T) {
		// Check before
		require.Equal(t, 42, e.Get((*Component1)(nil)).(*Component1).Num)

		e.Replace(&Component1{Num: 1234})

		// Check after replace
		require.Equal(t, 1234, e.Get((*Component1)(nil)).(*Component1).Num)
	})

	t.Run("Replace 2", func(t *testing.T) {
		e.Replace(&Component1{Num: 9876})

		// Check after replace
		require.Equal(t, 9876, e.Get((*Component1)(nil)).(*Component1).Num)
	})

	t.Run("Add component", func(t *testing.T) {
		// Check before replace
		require.Nil(t, e.Get((*Component2)(nil)))

		e.Replace(&Component2{Text: "Hello world"})

		// Check after replace
		require.Equal(t, "Hello world", e.Get((*Component2)(nil)).(*Component2).Text)
	})
}

func TestEntity_Delete(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	e.Replace(&Component1{Num: 42})
	e.Replace(&Component2{Text: "Hello world"})
	require.False(t, e.(*entity).destroyed)

	e.Delete((*Component1)(nil))
	require.Nil(t, e.Get((*Component1)(nil)))
	require.NotNil(t, e.Get((*Component2)(nil)))
	require.False(t, e.(*entity).destroyed)

	e.Delete((*Component2)(nil))
	require.Nil(t, e.Get((*Component1)(nil)))
	require.Nil(t, e.Get((*Component2)(nil)))
	require.True(t, e.(*entity).destroyed)

	e.Replace(&Component1{Num: 123})
	require.NotNil(t, e.Get((*Component1)(nil)))
	require.False(t, e.(*entity).destroyed)
}

func TestEntity_Components(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	c1 := &Component1{Num: 42}
	e.Replace(c1)
	c2 := &Component2{Text: "Hello world"}
	e.Replace(c2)

	cs := e.Components()
	require.Len(t, cs, 2)
	require.ElementsMatch(t, []Component{c1, c2}, cs)

	e.Delete((*Component1)(nil))

	cs = e.Components()
	require.Len(t, cs, 1)
	require.Equal(t, c2, cs[0])
}
