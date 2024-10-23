package tools

import (
	"errors"
	"math"
	"sync"
)

type ObjectManager[O comparable] struct {
	idLength   uint32
	idAlphabet string
	cap        int

	ids     map[string]O
	objects map[O]string
	mutex   sync.RWMutex
}

func NewObjectManager[O comparable](idLength uint32, idAlphabet string) (*ObjectManager[O], error) {
	if idLength < 1 {
		return nil, errors.New("idLength must be greater than 0")
	}
	if len(idAlphabet) < 2 {
		return nil, errors.New("idAlphabet must contain at least 2 characters")
	}

	return &ObjectManager[O]{
		idLength:   idLength,
		idAlphabet: idAlphabet,
		cap:        int(math.Pow(float64(len(idAlphabet)), float64(idLength)) * 0.9),

		ids:     make(map[string]O),
		objects: make(map[O]string),
	}, nil
}

func (manager *ObjectManager[D]) Add(object D) (string, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if len(manager.objects) >= manager.cap {
		return "", errors.New("maximum number of entries reached")
	}

	id := GenerateRandomString(manager.idLength, manager.idAlphabet)
	for _, ok := manager.ids[id]; ok; id = GenerateRandomString(manager.idLength, manager.idAlphabet) {
	}

	manager.ids[id] = object
	manager.objects[object] = id

	return id, nil
}

func (manager *ObjectManager[D]) RemoveId(id string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	object, ok := manager.ids[id]
	if !ok {
		return errors.New("entry not found")
	}

	delete(manager.ids, id)
	delete(manager.objects, object)

	return nil
}

func (manager *ObjectManager[D]) Remove(object D) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	id, ok := manager.objects[object]
	if !ok {
		return errors.New("entry not found")
	}
	delete(manager.ids, id)
	delete(manager.objects, object)

	return nil
}

func (manager *ObjectManager[D]) Get(id string) D {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	if object, ok := manager.ids[id]; ok {
		return object
	}

	var nilValue D
	return nilValue
}

func (manager *ObjectManager[D]) GetId(object D) string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	if id, ok := manager.objects[object]; ok {
		return id
	}

	return ""
}
