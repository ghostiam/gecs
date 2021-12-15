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

	w.Run(600)
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
		{[]ecs.Component{(*InputEvent)(nil)}, nil},
		{[]ecs.Component{(*Player)(nil), (*Position)(nil)}, nil},
	}
}

func (s *MovePlayerSystem) Update(dt float32, filtered [][]ecs.Entity) {
	input := filtered[0]
	player := filtered[1]

	if len(input) == 0 {
		return
	}

	fmt.Println(input)
	fmt.Println(player)

	ie := input[0].Get((*InputEvent)(nil)).(*InputEvent)
	pos := player[0].Get((*Position)(nil)).(*Position)
	pos.X += ie.Horizontal
	pos.Y -= ie.Vertical
}

type RenderConsoleSystem struct {
}

func (s *RenderConsoleSystem) GetFilters() []ecs.SystemFilter {
	return []ecs.SystemFilter{
		{[]ecs.Component{(*Position)(nil), (*RenderConsole)(nil)}, nil},
	}
}

func (s *RenderConsoleSystem) Update(dt float32, filtered [][]ecs.Entity) {
	term.Sync()
	term.Clear(term.ColorDefault, term.ColorDefault)

	for _, e := range filtered[0] {
		pos := e.Get((*Position)(nil)).(*Position)
		char := e.Get((*RenderConsole)(nil)).(*RenderConsole).Char

		fmt.Println(pos.X, pos.Y, string(char))
		term.SetCell(pos.X, pos.Y, char, term.ColorDefault, term.ColorGreen)
	}
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
