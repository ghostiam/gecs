package ecs

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestEcs(t *testing.T) {
	w := NewWorld()

	w.AddSystem(&Component1System{})
	w.AddSystem(&Component2System{})
	// w.AddSystem(&Component1And2System{})

	{
		e := w.NewEntity()
		c1 := &Component1{Num: 1111}
		e.Get(c1)
	}

	{
		e := w.NewEntity()
		c2 := &Component2{Text: "e2"}
		e.Replace(c2)
	}

	{
		e := w.NewEntity()
		c1 := &Component1{Num: 3333}
		e.Get(c1)

		c2 := &Component2{Text: "e3"}
		e.Replace(c2)

		get1 := e.Get((*Component1)(nil))
		get2 := e.Get((*Component2)(nil))

		spew.Dump(get1, get2)

		// e.Delete((*Component2)(nil))
	}

	w.AddSystem(&Component1And2System{})

	// e.Delete(&Component1{})
	// e.Destroy()

	spew.Dump(w)

	last := time.Now()
	// for {
	dt := time.Since(last)
	w.Update(float32(dt.Seconds()))
	// last = time.Now()

	// time.Sleep(500 * time.Millisecond)
	// }
}
