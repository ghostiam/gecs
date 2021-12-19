package gecs

import (
	"reflect"
)

// SystemFilter
// Include - компоненты которые должны быть на entity.
// Exclude - компоненты которых не должно быть на entity.
type SystemFilter struct {
	Include []Component
	Exclude []Component
}

type System interface {
	// GetFilters возвращает фильтры со списком компонентов.
	GetFilters() []SystemFilter

	// Update вызывается при каждом тике.
	// dt - время в секундах, прошедшее с предыдущего тика.
	// filtered - отфильтрованный список entity по фильтрам из метода GetFilters.
	// filtered - [FilterIndex][EntityIndex]Entity
	Update(dt float32, filtered [][]Entity)
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

	if len(componentTypes) == 0 {
		w.systemCacheDeleteEntityFromAllSystems(e)
		return
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
			ffid := filterIndex(fid)

			if hasComponentCount(f.Exclude) > 0 {
				w.systemCacheDeleteEntityFromSystem(e, st, ffid)
				continue
			}

			if hasComponentCount(f.Include) == len(f.Include) {
				if w.systemFiltersEntityCache[st] == nil {
					w.systemFiltersEntityCache[st] = make(map[filterIndex][]Entity)
				}

				w.systemFiltersEntityCache[st][ffid] = appendIfMissing(w.systemFiltersEntityCache[st][ffid], e)
				continue
			}

			w.systemCacheDeleteEntityFromSystem(e, st, ffid)
		}
	}
}

func (w *world) systemEntityCacheRebuildBySystem(systemType reflect.Type) {
	filter := w.systemFilters[systemType]
	if len(filter) == 0 {
		return
	}

	for fid, f := range filter {
		ffid := filterIndex(fid)

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
				return // If there is not at least one component from include, there is no point in further checking.
			}

			for e := range w.components[in] {
				if _, exist := excludeIDs[e.ID()]; exist {
					continue
				}

				includeIDs[e.ID()]++

				if includeIDs[e.ID()] != len(f.Include) {
					continue
				}

				// Append if system includes count  == entity component count
				if w.systemFiltersEntityCache[systemType] == nil {
					w.systemFiltersEntityCache[systemType] = make(map[filterIndex][]Entity)
				}

				w.systemFiltersEntityCache[systemType][ffid] = appendIfMissing(w.systemFiltersEntityCache[systemType][ffid], e)
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
