package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) AsyncMessage(topic, payload string, clientNames ...string) error {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return Error.New("Server stopped", nil)
	}
	server.mutex.Lock()
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				if server.errorLogger != nil {
					server.errorLogger.Log(Error.New("Client \""+clientName+"\" not found", nil).Error())
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
	server.statusMutex.RUnlock()

	errChannel := SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
	go func() {
		for err := range errChannel {
			if server.errorLogger != nil {
				server.errorLogger.Log(err.Error())
			}
		}
	}()
	return nil
}

func (server *SystemgeServer) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, error) {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil, Error.New("Server stopped", nil)
	}
	server.mutex.Lock()
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				if server.errorLogger != nil {
					server.errorLogger.Log(Error.New("Client \""+clientName+"\" not found", nil).Error())
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.mutex.Unlock()
	server.statusMutex.RUnlock()

	responseChannel, errChannel := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	go func() {
		for err := range errChannel {
			if server.errorLogger != nil {
				server.errorLogger.Log(err.Error())
			}
		}
	}()
	return responseChannel, nil
}
