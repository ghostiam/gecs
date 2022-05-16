package ecs

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/ghostiam/gecs"
)

type InputSystem struct {
	w gecs.World

	lastInput Vector2
}

func (s *InputSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*Player)(nil)}},
	}
}

func (s *InputSystem) Update(_ time.Duration, filtered [][]gecs.Entity) {
	players := filtered[0]
	if len(players) == 0 {
		return
	}

	player := players[0]

	var updated bool
	sdl.Do(func() {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				s.w.Stop()
				return
			case *sdl.KeyboardEvent:
				// spew.Dump(e)

				if e.Repeat > 0 {
					continue
				}

				var setVal int
				if e.Type == sdl.KEYUP {
					setVal = 0
					updated = true
				}
				if e.Type == sdl.KEYDOWN {
					setVal = 1
					updated = true
				}

				switch e.Keysym.Scancode {
				case sdl.SCANCODE_UP:
					s.lastInput.Y = -setVal
				case sdl.SCANCODE_DOWN:
					s.lastInput.Y = setVal
				case sdl.SCANCODE_LEFT:
					s.lastInput.X = -setVal
				case sdl.SCANCODE_RIGHT:
					s.lastInput.X = setVal
				}
			default:
				// spew.Dump(e)
			}
		}
	})

	if !updated && s.lastInput == (Vector2{}) {
		return
	}

	player.Replace(&InputEvent{Vector2: s.lastInput})
}

type MoveSystem struct {
	Velocity int
	Bounds   Bounds
}

func (s *MoveSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*InputEvent)(nil), (*Player)(nil), (*Position)(nil)}},
	}
}

func (s *MoveSystem) Update(delta time.Duration, filtered [][]gecs.Entity) {
	dts := delta.Seconds()
	players := filtered[0]
	for _, p := range players {
		input := p.Get((*InputEvent)(nil)).(*InputEvent)
		pos := p.Get((*Position)(nil)).(*Position)
		pos.X += int(float64(s.Velocity)*dts) * input.X
		pos.Y += int(float64(s.Velocity)*dts) * input.Y

		if s.Bounds.X >= pos.X {
			pos.X = s.Bounds.X
		}
		if s.Bounds.Y >= pos.Y {
			pos.Y = s.Bounds.Y
		}

		if s.Bounds.Width <= pos.X {
			pos.X = s.Bounds.Width
		}
		if s.Bounds.Height <= pos.Y {
			pos.Y = s.Bounds.Height
		}
	}
}

type CollideSystem struct{}

func (s *CollideSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*BoxCollider)(nil), (*Position)(nil)}},
		{Include: []gecs.Component{(*CircleCollider)(nil), (*Position)(nil)}},
	}
}

func (s *CollideSystem) Update(_ time.Duration, filtered [][]gecs.Entity) {
	boxes := filtered[0]
	circles := filtered[1]

	for _, a := range boxes {
		aBox := a.Get((*BoxCollider)(nil)).(*BoxCollider)
		aPos := a.Get((*Position)(nil)).(*Position)
		aRect := aBox.ToRect(aPos)
		ace := a.Get(&CollideEvent{}).(*CollideEvent)

	boxesLoop:
		for _, b := range boxes {
			if a.ID() == b.ID() {
				continue
			}

			for _, exist := range ace.Entities {
				if exist.ID() == b.ID() {
					continue boxesLoop
				}
			}

			bBox := b.Get((*BoxCollider)(nil)).(*BoxCollider)
			bPos := b.Get((*Position)(nil)).(*Position)
			bRect := bBox.ToRect(bPos)

			if aRect.HasIntersection(bRect) {
				ace.Entities = append(ace.Entities, b)

				bce := b.Get(&CollideEvent{}).(*CollideEvent)
				bce.Entities = append(bce.Entities, a)
			}
		}
	}

	for _, a := range circles {
		aBox := a.Get((*CircleCollider)(nil)).(*CircleCollider)
		aPos := a.Get((*Position)(nil)).(*Position)
		aCircle := aBox.ToCircle(aPos)
		ace := a.Get(&CollideEvent{}).(*CollideEvent)

	circlesLoop:
		for _, b := range boxes {
			if a.ID() == b.ID() {
				continue
			}

			for _, exist := range ace.Entities {
				if exist.ID() == b.ID() {
					continue circlesLoop
				}
			}

			bBox := b.Get((*BoxCollider)(nil)).(*BoxCollider)
			bPos := b.Get((*Position)(nil)).(*Position)
			bRect := bBox.ToRect(bPos)

			if aCircle.HasIntersectionWithRect(bRect) {
				ace.Entities = append(ace.Entities, b)

				bce := b.Get(&CollideEvent{}).(*CollideEvent)
				bce.Entities = append(bce.Entities, a)
			}
		}
	}
}

type CollectSystem struct{}

func (s *CollectSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*Collectable)(nil), (*CollideEvent)(nil)}},
	}
}

func (s *CollectSystem) Update(_ time.Duration, filtered [][]gecs.Entity) {
	for _, c := range filtered[0] {
		cc := c.Get((*CollideEvent)(nil)).(*CollideEvent)
		for _, e := range cc.Entities {
			if e.Has((*Player)(nil)) {
				c.Destroy()
				break
			}
		}
	}
}

type RenderSystem struct {
	Title string
	Size

	window   *sdl.Window
	renderer *sdl.Renderer

	dtSum      time.Duration
	frameCount int
	lastFPS    float64
}

func (s *RenderSystem) Init() error {
	var err error

	sdl.Do(func() {
		s.window, err = sdl.CreateWindow(s.Title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(s.Width), int32(s.Height), sdl.WINDOW_SHOWN)
		if err != nil {
			return
		}

		s.renderer, err = sdl.CreateRenderer(s.window, -1, sdl.RENDERER_SOFTWARE)
		if err != nil {
			return
		}
	})

	return err
}

func (s *RenderSystem) Destroy() {
	sdl.Do(func() {
		_ = s.renderer.Destroy()
		_ = s.window.Destroy()
		sdl.Quit()
	})
}

func (s *RenderSystem) GetFilters() []gecs.SystemFilter {
	return []gecs.SystemFilter{
		{Include: []gecs.Component{(*Position)(nil), (*RenderBox)(nil)}},
		{Include: []gecs.Component{(*Position)(nil), (*RenderCircle)(nil)}},
	}
}

func (s *RenderSystem) Update(delta time.Duration, filtered [][]gecs.Entity) {
	boxes := filtered[0]
	circles := filtered[1]

	s.dtSum += delta
	s.frameCount++

	if s.dtSum >= time.Second {
		s.lastFPS = float64(s.frameCount) * s.dtSum.Seconds()

		s.dtSum = 0
		s.frameCount = 0

		sdl.Do(func() {
			s.window.SetTitle(fmt.Sprintf("%s (FPS: %.2f)", s.Title, s.lastFPS))
		})
	}

	sdl.Do(func() {
		width, height := s.window.GetSize()

		// Clean screen
		_ = s.renderer.Clear()
		_ = s.renderer.SetDrawColor(0, 0, 0, 0x20)
		_ = s.renderer.FillRect(&sdl.Rect{0, 0, width, height})

		for _, c := range boxes {
			pos := c.Get((*Position)(nil)).(*Position)
			r := c.Get((*RenderBox)(nil)).(*RenderBox)
			cc, ok := c.Get((*CollideEvent)(nil)).(*CollideEvent)

			if ok && len(cc.Entities) > 0 {
				_ = s.renderer.SetDrawColor(0, 0, 255, 255)
			} else {
				_ = s.renderer.SetDrawColor(r.R, r.G, r.B, r.A)
			}
			_ = s.renderer.FillRect(&sdl.Rect{int32(pos.X - r.Width/2), int32(pos.Y - r.Height/2), int32(r.Width), int32(r.Height)})
		}

		for _, c := range circles {
			pos := c.Get((*Position)(nil)).(*Position)
			r := c.Get((*RenderCircle)(nil)).(*RenderCircle)
			cc, ok := c.Get((*CollideEvent)(nil)).(*CollideEvent)

			if ok && len(cc.Entities) > 0 {
				_ = s.renderer.SetDrawColor(0, 255, 255, 255)
			} else {
				_ = s.renderer.SetDrawColor(r.R, r.G, r.B, r.A)
			}
			FillCircle(s.renderer, int32(pos.X), int32(pos.Y), int32(r.Radius))
		}

		s.renderer.Present()
	})
}
