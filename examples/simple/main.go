package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/ghostiam/gecs"
)

func main() {
	const width, height = 60, 10
	rnd := rand.New(rand.NewSource(time.Now().Unix()))

	w := gecs.NewWorld()

	// Add systems
	w.AddSystem(&RandomMoveSystem{Rnd: rnd, MaxX: width, MaxY: height})
	w.AddSystem(&TextRenderSystem{BorderChar: '#', Width: width, Height: height})

	// Create entities
	for i := 0; i < 10; i++ {
		e := w.NewEntity()
		e.Get(&Position{X: rnd.Intn(width), Y: rnd.Intn(height)})
		e.Get(&TextRender{Char: rune('a' + i)})
	}

	w.Run(1)
}

type Position struct {
	X, Y int
}

type TextRender struct {
	Char rune
}

type RandomMoveSystem struct {
	Rnd  *rand.Rand
	MaxX int
	MaxY int
}

func (s *RandomMoveSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*Position)(nil)}},
	}
}

func (s *RandomMoveSystem) Update(_ time.Duration, filtered [][]gecs.Entity) {
	positions := filtered[0]

	for _, p := range positions {
		pos := p.Get((*Position)(nil)).(*Position)
		pos.X += s.Rnd.Intn(3) - 1
		pos.Y -= s.Rnd.Intn(3) - 1

		// Check bounds
		if pos.X < 0 {
			pos.X = 0
		}
		if pos.X >= s.MaxX {
			pos.X = s.MaxX - 1
		}

		if pos.Y < 0 {
			pos.Y = 0
		}
		if pos.Y >= s.MaxY {
			pos.Y = s.MaxY - 1
		}
	}
}

type TextRenderSystem struct {
	BorderChar    rune
	Width, Height int
}

func (s *TextRenderSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*Position)(nil), (*TextRender)(nil)}},
	}
}

func (s *TextRenderSystem) Update(_ time.Duration, filtered [][]gecs.Entity) {
	width, height := s.Width, s.Height
	hasBorder := s.BorderChar != rune(0)
	if hasBorder {
		width += 2  // left and right
		height += 2 // top and bottom
	}

	// Get positions and rune
	posRune := make(map[int]map[int]rune)
	for _, entity := range filtered[0] {
		pos := entity.Get((*Position)(nil)).(*Position)
		render := entity.Get((*TextRender)(nil)).(*TextRender)

		var offset int
		if hasBorder {
			offset = 1
		}

		x := pos.X + offset
		y := pos.Y + offset
		posRune[x] = map[int]rune{y: render.Char}
	}

	// Render in buffer
	var buf strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			isBorder := y == 0 || y == height-1 || x == 0 || x == width-1
			if isBorder {
				buf.WriteRune(s.BorderChar)
				continue
			}

			r, ok := posRune[x][y]
			if !ok {
				r = ' '
			}

			buf.WriteRune(r)
		}
		buf.WriteRune('\n')
	}

	// Write to console
	fmt.Println(buf.String())
}
