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

// assigns unique id to object and stores it in the manager.
// id / object can be resolved by the other.
// returns id and error
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

func (manager *ObjectManager[D]) AddId(id string, object D) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, ok := manager.ids[id]; ok {
		return errors.New("id already exists")
	}

	manager.ids[id] = object
	manager.objects[object] = id

	return nil
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

func (manager *ObjectManager[D]) GetBulk(ids ...string) []D {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	if len(ids) == 0 {
		objects := make([]D, 0, len(manager.ids))
		for _, object := range manager.ids {
			objects = append(objects, object)
		}
		return objects
	} else {
		objects := make([]D, 0, len(ids))
		for _, id := range ids {
			if object, ok := manager.ids[id]; ok {
				objects = append(objects, object)
			}
		}
		return objects
	}
}

func (manager *ObjectManager[D]) GetBulkId(objects ...D) []string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	if len(objects) == 0 {
		ids := make([]string, 0, len(manager.objects))
		for _, id := range manager.objects {
			ids = append(ids, id)
		}
		return ids
	} else {
		ids := make([]string, 0, len(objects))
		for _, object := range objects {
			if id, ok := manager.objects[object]; ok {
				ids = append(ids, id)
			}
		}
		return ids
	}
}

func (manager *ObjectManager[D]) IdExists(id string) bool {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	_, ok := manager.ids[id]
	return ok
}

func (manager *ObjectManager[D]) ObjectExists(object D) bool {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	_, ok := manager.objects[object]
	return ok
}

func (manager *ObjectManager[D]) ReplaceObject(oldObject, newObject D) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	id, ok := manager.objects[oldObject]
	if !ok {
		return errors.New("entry not found")
	}

	manager.ids[id] = newObject
	manager.objects[newObject] = id
	delete(manager.objects, oldObject)

	return nil
}

func (manager *ObjectManager[D]) ReplaceId(oldId, newId string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	object, ok := manager.ids[oldId]
	if !ok {
		return errors.New("entry not found")
	}

	manager.ids[newId] = object
	manager.objects[object] = newId
	delete(manager.ids, oldId)

	return nil
}

func (manager *ObjectManager[D]) GetSize() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return len(manager.ids)
}

func (manager *ObjectManager[D]) GetCapacity() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return manager.cap
}

func (manager *ObjectManager[D]) GetRemainingCapacity() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return manager.cap - len(manager.ids)
}
