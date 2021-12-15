package ecs

import (
	"reflect"
)

type World interface {
	NewEntity() Entity

	AddSystem(s System)
	RemoveSystem(s System)

	Update(dt float32)
}

func NewWorld() World {
	return &world{
		entityID:   0,
		entities:   nil,
		components: make(map[componentType]map[Entity]Component),

		systems:                  nil,
		systemFilters:            make(map[systemType][]systemFilterTypes),
		systemFiltersEntityCache: make(map[systemType]map[filterIndex]map[Entity]struct{}),
	}
}

// Type aliases for better readability.
type componentType reflect.Type
type systemType reflect.Type
type filterIndex int

type world struct {
	entityID uint64
	entities []Entity

	components map[componentType]map[Entity]Component

	systems                  []System
	systemFilters            map[systemType][]systemFilterTypes
	systemFiltersEntityCache map[systemType]map[filterIndex]map[Entity]struct{}
}

func (w *world) NewEntity() Entity {
	w.entityID++ // TODO atomic or mutex
	e := &entity{w: w, id: w.entityID}

	w.entities = append(w.entities, e)
	return e
}

func (w *world) Entities() []Entity {
	return w.entities
}

func (w *world) Systems() []System {
	return w.systems
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

	var deleteIdx = -1
	for i, ss := range w.systems {
		sst := reflect.TypeOf(ss)
		if st == sst {
			deleteIdx = i
			break
		}
	}

	if deleteIdx > -1 {
		w.systems = append(w.systems[:deleteIdx], w.systems[deleteIdx+1:]...)
	}

	delete(w.systemFiltersEntityCache, st)
	delete(w.systemFilters, st)
}

func (w *world) Update(dt float32) {
	for _, s := range w.systems {
		st := reflect.TypeOf(s)

		var filteredEntities [][]Entity
		if len(w.systemFiltersEntityCache[st]) > 0 {
			for fid := range w.systemFilters[st] {
				entities := make([]Entity, 0)
				for e := range w.systemFiltersEntityCache[st][filterIndex(fid)] {
					entities = append(entities, e)
				}

				filteredEntities = append(filteredEntities, entities)
			}
		}

		s.Update(dt, filteredEntities)
	}
}