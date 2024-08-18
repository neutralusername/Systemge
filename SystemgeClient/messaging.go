package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (client *SystemgeClient) AsyncMessage(topic, payload string, clientNames ...string) <-chan error {
	client.statusMutex.RLock()
	if client.status == Status.STOPPED {
		client.statusMutex.RUnlock()
		return nil
	}
	client.mutex.Lock()
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("Client \""+clientName+"\" not found", nil).Error())
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	return SystemgeConnection.MultiAsyncMessage(topic, payload, connections...)
}

func (client *SystemgeClient) SyncRequest(topic, payload string, clientNames ...string) (<-chan *Message.Message, <-chan error) {
	client.statusMutex.RLock()
	if client.status == Status.STOPPED {
		client.statusMutex.RUnlock()
		return nil, nil
	}
	client.mutex.Lock()
	connections := make([]*SystemgeConnection.SystemgeConnection, 0)
	if len(clientNames) == 0 {
		for _, connection := range client.nameConnections {
			connections = append(connections, connection)
		}
	} else {
		for _, clientName := range clientNames {
			connection := client.nameConnections[clientName]
			if connection == nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("Client \""+clientName+"\" not found", nil).Error())
				}
				continue
			}
			connections = append(connections, connection)
		}
	}
	client.mutex.Unlock()
	client.statusMutex.RUnlock()

	return SystemgeConnection.MultiSyncRequest(topic, payload, connections...)
}
