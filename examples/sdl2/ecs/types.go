package ecs

type Vector2 struct {
	X, Y int
}

type Size struct {
	Width, Height int
}

type Bounds struct {
	Vector2
	Size
}

type Rect = Bounds

func (a *Rect) Empty() bool {
	return a == nil || a.Width <= 0 || a.Height <= 0
}

func (a *Rect) HasIntersection(b *Rect) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Empty() || b.Empty() {
		return false
	}

	if a.X >= b.X+b.Width || a.X+a.Width <= b.X || a.Y >= b.Y+b.Height || a.Y+a.Height <= b.Y {
		return false
	}

	return true
}

type Circle struct {
	Vector2
	Radius int
}

func (a *Circle) Empty() bool {
	return a.Radius <= 0
}

// HasIntersectionWithRect https://stackoverflow.com/questions/401847/circle-rectangle-collision-detection-intersection
func (a *Circle) HasIntersectionWithRect(b *Rect) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Empty() || b.Empty() {
		return false
	}

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	rectXCenter := b.X + b.Width/2
	rectYCenter := b.Y + b.Height/2

	circleDistanceX := abs(a.X - rectXCenter)
	circleDistanceY := abs(a.Y - rectYCenter)

	if circleDistanceX > (b.Width/2 + a.Radius) {
		return false
	}
	if circleDistanceY > (b.Height/2 + a.Radius) {
		return false
	}

	if circleDistanceX <= b.Width/2 {
		return true
	}
	if circleDistanceY <= b.Height/2 {
		return true
	}

	crX := circleDistanceX - b.Width/2
	crY := circleDistanceY - b.Height/2
	cornerDistance := crX*crX + crY*crY
	return cornerDistance <= a.Radius*a.Radius
}
