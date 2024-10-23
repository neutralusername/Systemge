package systemge

import (
	"errors"
	"math"
	"sync"

	"github.com/neutralusername/systemge/tools"
)

type ConnectionManager[D any] struct {
	idLength      uint32
	idAlphabet    string
	connectionCap int

	mutex       sync.RWMutex
	ids         map[string]Connection[D]
	connections map[Connection[D]]string
}

func NewConnectionManager[D any](idLength uint32, idAlphabet string) (*ConnectionManager[D], error) {
	if idLength < 1 {
		return nil, errors.New("idLength must be greater than 0")
	}
	if len(idAlphabet) < 2 {
		return nil, errors.New("idAlphabet must contain at least 2 characters")
	}

	return &ConnectionManager[D]{
		idLength:      idLength,
		idAlphabet:    idAlphabet,
		connectionCap: int(math.Pow(float64(len(idAlphabet)), float64(idLength)) * 0.9),

		ids:         make(map[string]Connection[D]),
		connections: make(map[Connection[D]]string),
	}, nil
}

func (manager *ConnectionManager[D]) AddConnection(connection Connection[D]) (string, error) {
	if len(manager.connections) >= manager.connectionCap {
		manager.mutex.Unlock()
		return "", errors.New("maximum number of connections reached")
	}

	id := tools.GenerateRandomString(manager.idLength, manager.idAlphabet)
	for {
		if _, ok := manager.ids[id]; !ok {
			break
		}
		id = tools.GenerateRandomString(manager.idLength, manager.idAlphabet)
	}

	manager.ids[id] = connection
	manager.connections[connection] = id
	manager.mutex.Unlock()

	return id, nil
}

func (manager *ConnectionManager[D]) RemoveConnection(id string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	connection, ok := manager.ids[id]
	if !ok {
		return errors.New("connection not found")
	}
	delete(manager.ids, id)
	delete(manager.connections, connection)
	return nil
}

func (manager *ConnectionManager[D]) GetConnection(id string) Connection[D] {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	if connection, ok := manager.ids[id]; ok {
		return connection
	}
	return nil
}

func (manager *ConnectionManager[D]) GetId(connection Connection[D]) string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	if id, ok := manager.connections[connection]; ok {
		return id
	}
	return ""
}
