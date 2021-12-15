package main

import (
	"fmt"

	term "github.com/nsf/termbox-go"

	"github.com/ghostiam/ecs_mud/ecs"
)

func main() {
	err := term.Init()
	if err != nil {
		panic(err)
	}

	defer term.Close()

	w := ecs.NewWorld()
	w.AddSystem(NewInputConsoleSystem(w))
	w.AddSystem(&MovePlayerSystem{})
	w.AddSystem(&RenderConsoleSystem{})
	w.AddSystem(ecs.NewOneFrame((*InputEvent)(nil)))

	{
		playerEntity := w.NewEntity()
		playerEntity.Get(&Player{})
		playerEntity.Get(&Position{})
		playerEntity.Get(&RenderConsole{Char: 'w'})
	}

	// go func() {
	// 	for {
	// 		time.Sleep(500 * time.Millisecond)
	// 		w.NewEntity().Get(&InputEvent{Horizontal: 1})
	// 		time.Sleep(500 * time.Millisecond)
	// 		w.NewEntity().Get(&InputEvent{Vertical: -1})
	// 	}
	// }()

	w.Run(30)
}

func NewInputConsoleSystem(w ecs.World) *InputConsoleSystem {
	s := &InputConsoleSystem{
		w:           w,
		lastPressed: make(chan term.Key),
	}
	go s.readConsoleInput()
	return s
}

type InputConsoleSystem struct {
	w           ecs.World
	lastPressed chan term.Key
}

func (s *InputConsoleSystem) GetFilters() []ecs.SystemFilter {
	return nil
}

func (s *InputConsoleSystem) Update(dt float32, filtered [][]ecs.Entity) {
	event := InputEvent{}

	select {
	case key := <-s.lastPressed:
		// nolint: exhaustive
		switch key {
		case term.KeyCtrlC, term.KeyEsc:
			s.w.Stop()

		case term.KeyArrowUp:
			event.Vertical = 1
		case term.KeyArrowDown:
			event.Vertical = -1
		case term.KeyArrowLeft:
			event.Horizontal = -1
		case term.KeyArrowRight:
			event.Horizontal = 1
		}

		if event.Vertical == 0 && event.Horizontal == 0 {
			return
		}

		fmt.Println(event)
		s.w.NewEntity().Get(&event)
	default:
	}
}

func (s *InputConsoleSystem) readConsoleInput() {
	for {
		// nolint: exhaustive
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			s.lastPressed <- ev.Key
		case term.EventError:
			panic(ev.Err)
		}
	}
}

type MovePlayerSystem struct {
}

func (s *MovePlayerSystem) GetFilters() []ecs.SystemFilter {
	return []ecs.SystemFilter{
		{Include: []ecs.Component{(*InputEvent)(nil)}},
		{Include: []ecs.Component{(*Player)(nil), (*Position)(nil)}},
	}
}

func (s *MovePlayerSystem) Update(dt float32, filtered [][]ecs.Entity) {
	input := filtered[0]
	player := filtered[1]

	if len(input) == 0 {
		return
	}

	ie := input[0].Get((*InputEvent)(nil)).(*InputEvent)
	pos := player[0].Get((*Position)(nil)).(*Position)
	pos.X += ie.Horizontal
	pos.Y -= ie.Vertical
}

type RenderConsoleSystem struct {
	notFirstRun bool
}

func (s *RenderConsoleSystem) GetFilters() []ecs.SystemFilter {
	return []ecs.SystemFilter{
		{Include: []ecs.Component{(*Position)(nil), (*RenderConsole)(nil)}},
		{Include: []ecs.Component{(*InputEvent)(nil)}},
	}
}

func (s *RenderConsoleSystem) Update(dt float32, filtered [][]ecs.Entity) {
	if len(filtered[1]) == 0 && s.notFirstRun {
		return
	}
	s.notFirstRun = true

	_ = term.Clear(term.ColorDefault, term.ColorDefault)

	for _, e := range filtered[0] {
		pos := e.Get((*Position)(nil)).(*Position)
		char := e.Get((*RenderConsole)(nil)).(*RenderConsole).Char

		// fmt.Println(pos.X, pos.Y, string(char))
		term.SetCell(pos.X, pos.Y, char, term.ColorDefault, term.ColorGreen)

		for i, c := range fmt.Sprintf("(%d;%d)", pos.X, pos.Y) {
			term.SetCell(pos.X+i+1, pos.Y, c, term.ColorDefault, term.ColorGreen)
		}
	}

	_ = term.Sync()
}

type InputEvent struct {
	Vertical   int
	Horizontal int
}

type Player struct {
}

type Position struct {
	X, Y int
}

type RenderConsole struct {
	Char rune
}
