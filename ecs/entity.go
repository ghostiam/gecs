package ecs

import (
	"reflect"
)

type Component interface{}

type Entity interface {
	ID() uint64

	// Destroy удаляет все компоненты и удаляет entity из мира.
	// В случае, если кто-то держит ссылку на entity и добавит новый компонент, entity восстановится.
	Destroy()

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
	w              *world
	id             uint64
	componentCount uint64
	destroyed      bool
}

func (e *entity) ID() uint64 {
	return e.id
}

func (e *entity) Destroy() {
	e.destroyed = true

	var deleteIdx = -1
	for i, ee := range e.w.entities {
		if ee.ID() == e.ID() {
			deleteIdx = i
			break
		}
	}

	if deleteIdx > -1 {
		e.w.entities = append(e.w.entities[:deleteIdx], e.w.entities[deleteIdx+1:]...)
	}

	for ct, m := range e.w.components {
		delete(m, e)

		if len(e.w.components[ct]) == 0 {
			delete(e.w.components, ct)
		}
	}

	e.w.systemCacheDeleteEntityFromAllSystems(e)
}

func (e *entity) Get(c Component) Component {
	return e.getOrReplace(c, false)
}

func (e *entity) Has(c Component) bool {
	cs := e.w.components[reflect.TypeOf(c)]
	if cs == nil {
		return false
	}

	_, ok := cs[e]
	return ok
}

func (e *entity) Replace(c Component) {
	e.getOrReplace(c, true)
}

func (e *entity) Delete(c Component) {
	ct := reflect.TypeOf(c)
	delete(e.w.components[ct], e)
	if len(e.w.components[ct]) == 0 {
		delete(e.w.components, ct)
	}

	e.componentCount--
	if e.componentCount == 0 {
		e.Destroy()
		return
	}

	e.w.systemCacheRebuildByEntity(e)
}

func (e *entity) Components() []Component {
	var cs []Component

	for _, ec := range e.w.components {
		cs = append(cs, ec[e])
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
		cs = make(map[Entity]Component)
		e.w.components[ct] = cs
	}

	if !replace {
		v, ok := cs[e]
		if ok {
			return v
		}
	}

	if reflect.ValueOf(c).IsNil() {
		return nil
	}

	if e.destroyed {
		e.w.entities = append(e.w.entities, e)
		e.destroyed = false
	}

	e.componentCount++
	e.w.components[reflect.TypeOf(c)][e] = c
	e.w.systemCacheRebuildByEntity(e)
	return c
}
