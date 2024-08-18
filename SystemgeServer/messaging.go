package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) AsyncMessage(topic, payload string, clientNames ...string) map[string]error {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil
	}
	server.clientsMutex.Lock()
	clientNamesNotFound := make([]string, 0)
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				clientNamesNotFound = append(clientNamesNotFound, clientName)
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.clientsMutex.Unlock()
	server.statusMutex.RUnlock()

	errs := SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
	for _, clientName := range clientNamesNotFound {
		errs[clientName] = Error.New("Client not found", nil)
	}
	return errs
}

func (server *SystemgeServer) SyncRequest(topic, payload string, clientNames ...string) (map[string]*Message.Message, map[string]error) {
	server.statusMutex.RLock()
	if server.status == Status.STOPPED {
		server.statusMutex.RUnlock()
		return nil, nil
	}
	server.clientsMutex.Lock()
	clientNamesNotFound := make([]string, 0)
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range server.clients {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := server.clients[clientName]
			if connection == nil {
				clientNamesNotFound = append(clientNamesNotFound, clientName)
				continue
			}
			connections = append(connections, connection)
		}
	}
	server.clientsMutex.Unlock()
	server.statusMutex.RUnlock()

	responses, errs := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	for _, clientName := range clientNamesNotFound {
		errs[clientName] = Error.New("Client not found", nil)
	}
	return responses, errs
}
