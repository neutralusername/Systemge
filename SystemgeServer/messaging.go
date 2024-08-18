package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) AsyncMessage(topic, payload string, clientNames ...string) <-chan error {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil
	}
	server.clientsMutex.Lock()
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
	server.clientsMutex.Unlock()
	server.statusMutex.RUnlock()

	return SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
}

func (server *SystemgeServer) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, <-chan error) {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil, nil
	}
	server.clientsMutex.Lock()
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
	server.clientsMutex.Unlock()
	server.statusMutex.RUnlock()

	return SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
}
