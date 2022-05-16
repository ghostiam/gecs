package ecs

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

// FillCircle https://gist.github.com/derofim/912cfc9161269336f722
func FillCircle(r *sdl.Renderer, cx, cy, radius int32) {
	for dy := int32(1); dy <= radius; dy += 1 {
		// This loop is unrolled a bit, only iterating through half of the
		// height of the circle.  The result is used to draw a scan line and
		// its mirror image below it.

		// The following formula has been simplified from our original.  We
		// are using half of the width of the circle because we are provided
		// with a center and we need left/right coordinates.
		dx := int32(math.Floor(math.Sqrt(float64((2.0 * radius * dy) - (dy * dy)))))
		_ = r.DrawLine(cx-dx, cy+dy-radius, cx+dx, cy+dy-radius)
		_ = r.DrawLine(cx-dx, cy-dy+radius, cx+dx, cy-dy+radius)
	}
}

func DrawCircle(r *sdl.Renderer, cx, cy, radius int32) {
	diameter := radius * 2

	x := radius - 1
	y := int32(0)
	tx := int32(1)
	ty := int32(1)
	e := tx - diameter

	for x >= y {
		//  Each of the following renders an octant of the circle
		_ = r.DrawPoint(cx+x, cy-y)
		_ = r.DrawPoint(cx+x, cy+y)
		_ = r.DrawPoint(cx-x, cy-y)
		_ = r.DrawPoint(cx-x, cy+y)
		_ = r.DrawPoint(cx+y, cy-x)
		_ = r.DrawPoint(cx+y, cy+x)
		_ = r.DrawPoint(cx-y, cy-x)
		_ = r.DrawPoint(cx-y, cy+x)

		if e <= 0 {
			y++
			e += ty
			ty += 2
		}

		if e > 0 {
			x--
			tx += 2
			e += tx - diameter
		}
	}
}
