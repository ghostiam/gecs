package ecs

import (
	"reflect"
)

type World interface {
	NewEntity() EntityComponent
	DestroyEntity(e Entity)

	AddSystem(s System)
	RemoveSystem(s System)

	Update(dt float32)
}

func NewWorld() World {
	return &world{
		entityID:   0,
		entities:   nil,
		components: make(map[componentType]map[EntityID]Component),

		systems:                  nil,
		systemFilters:            make(map[systemType][]systemFilterTypes),
		systemFiltersEntityCache: make(map[systemType]map[int]map[EntityID]struct{}),
	}
}

// Type aliases for better readability.
type componentType reflect.Type
type systemType reflect.Type

type world struct {
	entityID EntityID
	entities []Entity

	// map[ComponentType]map[EntityID]Component
	components map[componentType]map[EntityID]Component
	// oneFrameComponents map[reflect.Type]map[uint64]Component // TODO

	systems []System
	// map[SystemType][]Filters
	systemFilters map[systemType][]systemFilterTypes
	// map[SystemType]map[FilterIndex]map[EntityID]struct{}
	systemFiltersEntityCache map[systemType]map[int]map[EntityID]struct{}
}

func (w *world) NewEntity() EntityComponent {
	w.entityID++ // TODO atomic or mutex
	e := &entity{w: w, id: w.entityID}

	w.entities = append(w.entities, e)
	return e
}

func (w *world) DestroyEntity(e Entity) {
	var deleteIdx = -1
	for i, ee := range w.entities {
		if ee.ID() == e.ID() {
			deleteIdx = i
			break
		}
	}

	if deleteIdx > -1 {
		w.entities = append(w.entities[:deleteIdx], w.entities[deleteIdx+1:]...)
	}

	for ct, m := range w.components {
		delete(m, e.ID())

		if len(w.components[ct]) == 0 {
			delete(w.components, ct)
		}
	}

	w.systemCacheDeleteEntityFromAllSystems(e)
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

		var filteredEntities [][]EntityComponent
		if len(w.systemFiltersEntityCache[st]) > 0 {
			for fid := range w.systemFilters[st] {
				entities := make([]EntityComponent, 0)
				for eid := range w.systemFiltersEntityCache[st][fid] {
					entities = append(entities, &entity{w: w, id: eid})
				}

				filteredEntities = append(filteredEntities, entities)
			}
		}

		s.Update(dt, filteredEntities)
	}
}
