package gecs

import (
	"reflect"
)

// Component ecs interface. Used for better readability.
type Component interface{}

// Entity ecs interface.
type Entity interface {
	ID() uint64

	// Destroy removes all components and removes the entity from the world.
	// In case someone holds a reference to the entity and adds a new component, the entity will be restored.
	Destroy()

	// Get gets an existing component with the type of the passed component.
	// If the component doesn't exist and is not nil passed, the component will be added to entity.
	Get(c Component) Component

	// Has returns true if there is a component with the passed type on entity.
	Has(c Component) bool

	// Replace adds a component to the entity if it doesn't exist, or replaces it if it exists.
	// If the type is nil, it does nothing.
	Replace(c Component)

	// Delete removes the component with the passed type.
	Delete(c Component)

	// Components returns all entity component.
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

	for i, ee := range e.w.entities {
		if ee.ID() == e.ID() {
			e.w.entities = append(e.w.entities[:i], e.w.entities[i+1:]...)
			break
		}
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

	_, ok := e.w.components[ct][e]
	if !ok {
		return
	}

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
