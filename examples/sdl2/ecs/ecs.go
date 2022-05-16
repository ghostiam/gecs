package ecs

import (
	"image/color"

	"github.com/ghostiam/gecs"
)

func Run() error {
	windowSize := Size{Width: 800, Height: 600}

	w := gecs.NewWorld()

	w.AddSystem(&InputSystem{w: w})
	w.AddSystem(&MoveSystem{Velocity: 300, Bounds: Bounds{Size: windowSize}})
	w.AddSystem(&CollideSystem{})
	w.AddSystem(&CollectSystem{})
	w.AddSystem(&RenderSystem{Title: "ECS example", Size: windowSize})
	w.AddSystem(gecs.NewOneFrame((*InputEvent)(nil)))
	w.AddSystem(gecs.NewOneFrame((*CollideEvent)(nil)))

	{
		player := w.NewEntity()
		player.Get(&Player{})
		player.Get(&Position{Vector2{windowSize.Width / 2, windowSize.Height / 2}})
		player.Get(&CircleCollider{Radius: 25})
		player.Get(&RenderCircle{Radius: 25, RGBA: color.RGBA{255, 0, 0, 255}})
	}

	for i := 0; i < 7; i++ {
		for j := 0; j < 5; j++ {
			e := w.NewEntity()
			e.Get(&Collectable{})
			e.Get(&Position{Vector2{100 + 100*i, 100 + 100*j}})
			e.Get(&BoxCollider{Size: Size{50, 50}})
			e.Get(&RenderBox{Size: Size{50, 50}, RGBA: color.RGBA{200, 200, 0, 255}})
		}
	}

	return w.Run(60)
}
