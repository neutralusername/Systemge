package systemge

import (
	"errors"
	"math"
	"sync"

	"github.com/neutralusername/systemge/tools"
)

type ConnectionManager[D any] struct {
	sessionIdLength   uint32
	sessionIdAlphabet string
	sessionCap        int

	sessionMutex sync.RWMutex
	sessions     map[string]Connection[D]
	connections  map[Connection[D]]string
}

func NewSessionManager[D any](sessionIdLength uint32, sessionIdAlphabet string) (*ConnectionManager[D], error) {
	if sessionIdLength < 1 {
		return nil, errors.New("sessionIdLength must be greater than 0")
	}
	if len(sessionIdAlphabet) < 2 {
		return nil, errors.New("sessionIdAlphabet must contain at least 2 characters")
	}
	return &ConnectionManager[D]{
		sessionIdLength:   sessionIdLength,
		sessionIdAlphabet: sessionIdAlphabet,
		sessionCap:        int(math.Pow(float64(len(sessionIdAlphabet)), float64(sessionIdLength)) * 0.9),

		sessions:    make(map[string]Connection[D]),
		connections: make(map[Connection[D]]string),
	}, nil
}

func (manager *ConnectionManager[D]) CreateSession(connection Connection[D]) (string, error) {
	if len(manager.sessions) >= manager.sessionCap {
		manager.sessionMutex.Unlock()
		return "", errors.New("maximum number of sessions reached")
	}

	id := tools.GenerateRandomString(manager.sessionIdLength, manager.sessionIdAlphabet)
	for {
		if _, ok := manager.sessions[id]; !ok {
			break
		}
		id = tools.GenerateRandomString(manager.sessionIdLength, manager.sessionIdAlphabet)
	}

	manager.sessions[id] = connection
	manager.connections[connection] = id
	manager.sessionMutex.Unlock()

	return id, nil
}

func (manager *ConnectionManager[D]) RemoveSession(sessionId string) error {
	manager.sessionMutex.Lock()
	defer manager.sessionMutex.Unlock()
	connection, ok := manager.sessions[sessionId]
	if !ok {
		return errors.New("session not found")
	}
	delete(manager.sessions, sessionId)
	delete(manager.connections, connection)
	return nil
}

func (manager *ConnectionManager[D]) GetConnection(id string) Connection[D] {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if session, ok := manager.sessions[id]; ok {
		return session
	}
	return nil
}

func (manager *ConnectionManager[D]) GetId(connection Connection[D]) string {
	manager.sessionMutex.RLock()
	defer manager.sessionMutex.RUnlock()
	if id, ok := manager.connections[connection]; ok {
		return id
	}
	return ""
}
