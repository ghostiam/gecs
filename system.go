package gecs

import (
	"reflect"
	"time"
)

// SystemFilter contains components for filtering entity when calling System.Update method on the system.
//    Include - the components that should be on the entity.
//    Exclude - the components of which should not be on the entity.
type SystemFilter struct {
	Include []Component
	Exclude []Component
}

// SystemIniter ecs interface.
type SystemIniter interface {
	System

	Init() error
}

// System ecs interface.
type System interface {
	// GetFilters returns filters with a list of components.
	GetFilters() []SystemFilter

	// Update is called on every tick.
	// delta - time elapsed from the previous tick.
	// filtered - filtered entity list by filters from the GetFilters method.
	// Always contains the same number of elements as the GetFilters method returns, in the same filter order.
	// filtered - [FilterIndex][EntityIndex]Entity
	Update(delta time.Duration, filtered [][]Entity)
}

// SystemDestroyer ecs interface.
type SystemDestroyer interface {
	System

	Destroy()
}

// SystemFull ecs interface.
type SystemFull interface {
	SystemIniter
	SystemDestroyer
}

// systemFilterTypes includes Component types
type systemFilterTypes struct {
	Include []reflect.Type
	Exclude []reflect.Type
}

func (w *world) systemCacheDeleteEntityFromAllSystems(e Entity) {
	for st, fids := range w.systemFiltersEntityCache {
		for fid := range fids {
			w.systemCacheDeleteEntityFromSystem(e, st, fid)
		}
	}
}

func (w *world) systemCacheDeleteEntityFromSystem(e Entity, systemType reflect.Type, fid filterIndex) {
	for i, ee := range w.systemFiltersEntityCache[systemType][fid] {
		if ee.ID() == e.ID() {
			w.systemFiltersEntityCache[systemType][fid] = append(w.systemFiltersEntityCache[systemType][fid][:i], w.systemFiltersEntityCache[systemType][fid][i+1:]...)
			break
		}
	}

	if len(w.systemFiltersEntityCache[systemType][fid]) == 0 {
		delete(w.systemFiltersEntityCache[systemType], fid)
	}

	if len(w.systemFiltersEntityCache[systemType]) == 0 {
		delete(w.systemFiltersEntityCache, systemType)
	}
}

func (w *world) systemCacheRebuildByEntity(e Entity) {
	var componentTypes []reflect.Type
	for ct, me := range w.components {
		_, ok := me[e]
		if !ok {
			continue
		}

		componentTypes = append(componentTypes, ct)
	}

	hasComponentCount := func(ts []reflect.Type) int {
		found := 0

		for _, c := range componentTypes {
			for _, t := range ts {
				if c == t {
					found++
				}
			}
		}

		return found
	}

	for _, s := range w.systems {
		st := reflect.TypeOf(s)

		filter := w.systemFilters[st]
		if len(filter) == 0 {
			continue
		}

		for fid, f := range filter {
			if hasComponentCount(f.Exclude) > 0 {
				w.systemCacheDeleteEntityFromSystem(e, st, fid)
				continue
			}

			if hasComponentCount(f.Include) == len(f.Include) {
				if w.systemFiltersEntityCache[st] == nil {
					w.systemFiltersEntityCache[st] = make(map[filterIndex][]Entity)
				}

				w.systemFiltersEntityCache[st][fid] = appendIfMissing(w.systemFiltersEntityCache[st][fid], e)
				continue
			}

			w.systemCacheDeleteEntityFromSystem(e, st, fid)
		}
	}
}

func (w *world) systemEntityCacheRebuildBySystem(systemType reflect.Type) {
	filter := w.systemFilters[systemType]
	if len(filter) == 0 {
		return
	}

	for fid, f := range filter {
		excludeIDs := make(map[uint64]struct{}) // map[EntityID]struct{}
		for _, ex := range f.Exclude {
			if len(w.components[ex]) == 0 {
				continue
			}

			for e := range w.components[ex] {
				excludeIDs[e.ID()] = struct{}{}
			}
		}

		includeIDs := make(map[uint64]int) // map[EntityID]count
		for _, in := range f.Include {
			if len(w.components[in]) == 0 {
				// TODO add coverage
				break // If there is not at least one component from include, there is no point in further checking.
			}

			for e := range w.components[in] {
				if _, exist := excludeIDs[e.ID()]; exist {
					continue
				}

				includeIDs[e.ID()]++

				if includeIDs[e.ID()] != len(f.Include) {
					continue
				}

				// Append if system includes count == entity component count
				if w.systemFiltersEntityCache[systemType] == nil {
					w.systemFiltersEntityCache[systemType] = make(map[filterIndex][]Entity)
				}

				w.systemFiltersEntityCache[systemType][fid] = appendIfMissing(w.systemFiltersEntityCache[systemType][fid], e)
			}
		}
	}
}

func appendIfMissing(slice []Entity, i Entity) []Entity {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}
