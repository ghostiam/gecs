package gecs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystem_NewOneFrame(t *testing.T) {
	w := NewWorld()
	w.AddSystem(NewOneFrame((*Component1)(nil)))

	e := w.NewEntity()
	e.Replace(&Component1{Num: 42})

	require.True(t, e.Has((*Component1)(nil)))

	w.Update(0.1)

	require.False(t, e.Has((*Component1)(nil)))
}
