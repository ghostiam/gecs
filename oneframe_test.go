package gecs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type OneFrameComponent struct {
	Event string
}

func TestSystem_NewOneFrame(t *testing.T) {
	w := NewWorld()
	w.AddSystem(NewOneFrame((*OneFrameComponent)(nil)))

	e := w.NewEntity()
	e.Replace(&OneFrameComponent{Event: "EventName"})

	require.True(t, e.Has((*OneFrameComponent)(nil)))

	w.SystemsUpdate(time.Second)

	require.False(t, e.Has((*OneFrameComponent)(nil)))
}
