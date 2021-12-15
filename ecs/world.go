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
		components: make(map[reflect.Type]map[uint64]Component),

		systems:           nil,
		systemEntityCache: make(map[reflect.Type]map[uint64]struct{}),
		systemIn:          make(map[reflect.Type][]reflect.Type),
		systemEx:          make(map[reflect.Type][]reflect.Type),
	}
}

type world struct {
	entityID uint64
	entities []Entity

	// map[ComponentType]map[EntityID]Component
	components map[reflect.Type]map[uint64]Component
	// oneFrameComponents map[reflect.Type]map[uint64]Component // TODO

	systems []System
	// map[SystemType]map[EntityID]struct{}
	systemEntityCache map[reflect.Type]map[uint64]struct{}
	// map[SystemType][]ComponentType
	systemIn map[reflect.Type][]reflect.Type
	systemEx map[reflect.Type][]reflect.Type
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

	in, ex := s.GetFilter()
	for _, v := range in {
		w.systemIn[st] = append(w.systemIn[st], reflect.TypeOf(v))
	}
	for _, v := range ex {
		w.systemEx[st] = append(w.systemEx[st], reflect.TypeOf(v))
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

	delete(w.systemEntityCache, st)
	delete(w.systemIn, st)
	delete(w.systemEx, st)
}

func (w *world) Update(dt float32) {
	for _, s := range w.systems {
		st := reflect.TypeOf(s)

		var entities []EntityComponent
		if len(w.systemEntityCache[st]) > 0 {
			entities = make([]EntityComponent, 0, len(w.systemEntityCache[st]))
			for eid := range w.systemEntityCache[st] {
				entities = append(entities, &entity{w: w, id: eid})
			}
		}

		s.Update(dt, entities)
	}
}
