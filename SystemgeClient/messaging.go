package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (client *SystemgeClient) AsyncMessage(topic, payload string, clientNames ...string) map[string]error {
	client.statusMutex.RLock()
	if client.status == Status.STOPPED {
		client.statusMutex.RUnlock()
		return nil
	}
	client.mutex.Lock()
	clientNamesNotFound := make([]string, 0)
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				clientNamesNotFound = append(clientNamesNotFound, clientName)
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	errs := SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
	for _, clientName := range clientNamesNotFound {
		errs[clientName] = Error.New("Client not found", nil)
	}
	return errs
}

func (client *SystemgeClient) SyncRequest(topic, payload string, clientNames ...string) (map[string]*Message.Message, map[string]error) {
	client.statusMutex.RLock()
	if client.status == Status.STOPPED {
		client.statusMutex.RUnlock()
		return nil, nil
	}
	client.mutex.Lock()
	clientNamesNotFound := make([]string, 0)
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				clientNamesNotFound = append(clientNamesNotFound, clientName)
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	responses, errs := SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
	for _, clientName := range clientNamesNotFound {
		errs[clientName] = Error.New("Client not found", nil)
	}
	return responses, errs
}
