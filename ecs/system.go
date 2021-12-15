package ecs

import (
	"reflect"
)

type System interface {
	// GetFilter возвращает список компонентов.
	// include - компоненты которые должны быть на entity.
	// exclude - компоненты которых не должно быть на entity.
	GetFilter() (include, exclude []Component)

	// Update вызывается при каждом тике.
	// dt - время в секундах, прошедшее с предыдущего тика.
	// filtered - отфильтрованный список entity по фильтрам из метода GetFilter.
	Update(dt float32, filtered []EntityComponent)
}

func (w *world) systemCacheDeleteEntityFromAllSystems(e Entity) {
	for st := range w.systemEntityCache {
		w.systemCacheDeleteEntityFromSystem(e, st)
	}
}

func (w *world) systemCacheDeleteEntityFromSystem(e Entity, systemType reflect.Type) {
	delete(w.systemEntityCache[systemType], e.ID())
	if len(w.systemEntityCache[systemType]) == 0 {
		delete(w.systemEntityCache, systemType)
	}
}

func (w *world) systemCacheRebuildByEntity(e Entity) {
	var componentTypes []reflect.Type
	for ct, me := range w.components {
		_, ok := me[e.ID()]
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
		if hasComponentCount(w.systemEx[st]) > 0 {
			w.systemCacheDeleteEntityFromSystem(e, st)
			continue
		}

		if hasComponentCount(w.systemIn[st]) == len(w.systemIn[st]) {
			if w.systemEntityCache[st] == nil {
				w.systemEntityCache[st] = make(map[uint64]struct{})
			}

			w.systemEntityCache[st][e.ID()] = struct{}{}
			continue
		}

		w.systemCacheDeleteEntityFromSystem(e, st)
	}
}

func (w *world) systemEntityCacheRebuildBySystem(systemType reflect.Type) {
	excludes := w.systemEx[systemType]
	includes := w.systemIn[systemType]

	excludeIDs := make(map[uint64]struct{}) // map[EntityID]struct{}
	for _, ex := range excludes {
		if len(w.components[ex]) == 0 {
			continue
		}

		for eid := range w.components[ex] {
			excludeIDs[eid] = struct{}{}
		}
	}

	includeIDs := make(map[uint64]int) // map[EntityID]count
	for _, in := range includes {
		if len(w.components[in]) == 0 {
			return // If there is not at least one component from include, there is no point in further checking.
		}

		for eid := range w.components[in] {
			if _, exist := excludeIDs[eid]; exist {
				continue
			}

			includeIDs[eid]++

			if includeIDs[eid] != len(includes) {
				continue
			}

			// Append if system includes count  == entity component count
			if w.systemEntityCache[systemType] == nil {
				w.systemEntityCache[systemType] = make(map[uint64]struct{})
			}

			w.systemEntityCache[systemType][eid] = struct{}{}
		}
	}
}
