package ecs

import (
	"reflect"
)

type EntityID uint64

type Component interface{}

type Entity interface {
	ID() EntityID

	// Get получает существующий компонент с типом переданного компонента.
	// Если компонента не существует и передан не nil, компонент будет добавлен к entity.
	Get(c Component) Component

	// Has возвращает true, если существует компонент с переданным типом на entity.
	Has(c Component) bool

	// Replace добавляет компонент к entity, если он не существует или заменяет его, если существует.
	// Если передан nil тип, ничего не делает.
	Replace(c Component)

	// Delete удаляет компонент с переданным типом.
	Delete(c Component)

	// Components возвращает все компоненты entity.
	Components() []Component
}

type entity struct {
	w  *world
	id EntityID
}

func (e *entity) ID() EntityID {
	return e.id
}

func (e *entity) Get(c Component) Component {
	return e.getOrReplace(c, false)
}

func (e *entity) Has(c Component) bool {
	cs := e.w.components[reflect.TypeOf(c)]
	if cs == nil {
		return false
	}

	_, ok := cs[e.id]
	return ok
}

func (e *entity) Replace(c Component) {
	e.getOrReplace(c, true)
}

func (e *entity) Delete(c Component) {
	ct := reflect.TypeOf(c)
	delete(e.w.components[ct], e.id)
	if len(e.w.components[ct]) == 0 {
		delete(e.w.components, ct)
	}

	e.w.systemCacheRebuildByEntity(e)
}

func (e *entity) Components() []Component {
	var cs []Component

	for _, ec := range e.w.components {
		cs = append(cs, ec[e.id])
	}

	return cs
}

func (e *entity) getOrReplace(c Component, replace bool) Component {
	if c == nil {
		return nil
	}

	ct := reflect.TypeOf(c)
	cs := e.w.components[ct]
	if cs == nil {
		cs = make(map[EntityID]Component)
		e.w.components[ct] = cs
	}

	if !replace {
		v, ok := cs[e.id]
		if ok {
			return v
		}
	}

	if reflect.ValueOf(c).IsNil() {
		return nil
	}

	e.w.components[reflect.TypeOf(c)][e.ID()] = c
	e.w.systemCacheRebuildByEntity(e)
	return c
}
