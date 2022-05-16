package main

import (
	"runtime"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/ghostiam/gecs/examples/sdl2/ecs"
)

func main() {
	runtime.LockOSThread()

	sdl.Main(func() {
		err := ecs.Run()
		if err != nil {
			panic(err)
		}
	})
}
