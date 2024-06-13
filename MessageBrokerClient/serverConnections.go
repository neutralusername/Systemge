package MessageBrokerClient

import "Systemge/Error"

func (client *Client) addServerConnection(serverConnection *serverConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.serverConnections[serverConnection.address] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.serverConnections[serverConnection.address] = serverConnection
	go client.handleServerMessages(serverConnection)
	go client.heartbeatLoop(serverConnection)
	return nil
}

func (client *Client) getServerConnection(address string) *serverConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.serverConnections[address]
}
