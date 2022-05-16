package ecs

import (
	"image/color"

	"github.com/ghostiam/gecs"
)

// Markers

type Player struct{}

type Collectable struct{}

// Components with data

// Base

type Position struct {
	Vector2
}

// Events

type InputEvent struct {
	Vector2
}

type CollideEvent struct {
	Entities []gecs.Entity
}

// Render

type RenderBox struct {
	Size
	color.RGBA
}

type RenderCircle struct {
	Radius int
	color.RGBA
}

// Colliders

type BoxCollider struct {
	Offset Vector2
	Size
}

func (a *BoxCollider) ToRect(p *Position) *Rect {
	if a == nil || p == nil {
		return nil
	}

	return &Rect{
		Vector2: Vector2{
			X: a.Offset.X - a.Width/2 + p.X,
			Y: a.Offset.Y - a.Height/2 + p.Y,
		},
		Size: Size{
			Width:  a.Width,
			Height: a.Height,
		},
	}
}

type CircleCollider struct {
	Offset Vector2
	Radius int
}

func (c *CircleCollider) ToCircle(p *Position) *Circle {
	if c == nil || p == nil {
		return nil
	}

	return &Circle{
		Vector2: Vector2{
			X: c.Offset.X + p.X,
			Y: c.Offset.Y + p.Y,
		},
		Radius: c.Radius,
	}

}
