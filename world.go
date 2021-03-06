package gecs

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// World ecs interface.
type World interface {
	NewEntity() Entity

	AddSystem(s System)
	RemoveSystem(s System)

	SystemsInit() error
	// SystemsUpdate calls an update on all systems. Takes in the time elapsed from the previous call.
	SystemsUpdate(delta time.Duration)
	SystemsDestroy()

	// Run calls the Update method with a TPS (Tick per second) rate. Blocking method!
	Run(tps uint) error
	Stop()
}

// NewWorld creates new ecs world instance.
func NewWorld() World {
	return &world{
		entityID:   0,
		entities:   nil,
		components: make(map[componentType]map[Entity]Component),

		systems:                  nil,
		systemFilters:            make(map[systemType][]systemFilterTypes),
		systemFiltersEntityCache: make(map[systemType]map[filterIndex][]Entity),

		done: make(chan struct{}, 1),
	}
}

// Type aliases for better readability.
type componentType = reflect.Type
type systemType = reflect.Type
type filterIndex = int

type world struct {
	entityID uint64
	entities []Entity

	components map[componentType]map[Entity]Component

	systems                  []System
	systemFilters            map[systemType][]systemFilterTypes
	systemFiltersEntityCache map[systemType]map[filterIndex][]Entity

	done   chan struct{}
	ticker *time.Ticker
}

func (w *world) NewEntity() Entity {
	w.entityID++
	e := &entity{w: w, id: w.entityID}

	w.entities = append(w.entities, e)
	return e
}

func (w *world) AddSystem(s System) {
	w.RemoveSystem(s)

	w.systems = append(w.systems, s)

	st := reflect.TypeOf(s)

	for _, f := range s.GetFilters() {
		var in, ex []reflect.Type
		for _, v := range f.Include {
			in = append(in, reflect.TypeOf(v))
		}
		for _, v := range f.Exclude {
			ex = append(ex, reflect.TypeOf(v))
		}

		w.systemFilters[st] = append(w.systemFilters[st], systemFilterTypes{
			Include: in,
			Exclude: ex,
		})
	}

	if len(w.components) == 0 {
		return
	}

	w.systemEntityCacheRebuildBySystem(st)
}

func (w *world) RemoveSystem(s System) {
	st := reflect.TypeOf(s)

	for i, ss := range w.systems {
		sst := reflect.TypeOf(ss)
		if st == sst {
			w.systems = append(w.systems[:i], w.systems[i+1:]...)
			break
		}
	}

	delete(w.systemFiltersEntityCache, st)
	delete(w.systemFilters, st)
}

func (w *world) SystemsInit() error {
	for _, s := range w.systems {
		ss, ok := s.(SystemIniter)
		if !ok {
			continue
		}

		err := ss.Init()
		if err != nil {
			st := reflect.TypeOf(s)

			return fmt.Errorf("%s: %w", st.String(), err)
		}
	}

	return nil
}

func (w *world) SystemsUpdate(delta time.Duration) {
	for _, s := range w.systems {
		st := reflect.TypeOf(s)

		var filteredEntities [][]Entity
		for fid := range w.systemFilters[st] {
			var entities []Entity
			if len(w.systemFiltersEntityCache[st]) > 0 {
				entities = w.systemFiltersEntityCache[st][fid]
			}
			filteredEntities = append(filteredEntities, entities)
		}

		s.Update(delta, filteredEntities)
	}
}

func (w *world) SystemsDestroy() {
	for _, s := range w.systems {
		ss, ok := s.(SystemDestroyer)
		if !ok {
			continue
		}

		ss.Destroy()
	}
}

func (w *world) Run(fps uint) error {
	if fps == 0 {
		return errors.New("fps must be greater than 0")
	}

	err := w.SystemsInit()
	if err != nil {
		return err
	}

	delay := time.Second / time.Duration(fps)
	w.ticker = time.NewTicker(delay)

	last := time.Now()

loop:
	for {
		delta := time.Since(last)
		last = time.Now()
		w.SystemsUpdate(delta)

		select {
		case <-w.ticker.C:
			continue
		case <-w.done:
			break loop
		}
	}

	w.SystemsDestroy()
	return nil
}

func (w *world) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}

	w.done <- struct{}{}
}
