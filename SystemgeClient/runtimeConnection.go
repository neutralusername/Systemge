package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpConnection"
)

// AddConnection adds an active connection to the client.
// if reconnectEndpointConfig is not nil, the connection will attempt to reconnect
func (client *SystemgeClient) AddConnection(connection *TcpConnection.TcpConnection, reconnectEndpointConfig *Config.TcpEndpoint) error {
	if connection == nil {
		return Error.New("connection is nil", nil)
	}
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status == Status.STOPPED {
		return Error.New("client stopped", nil)
	}
	if _, ok := client.addressConnections[connection.GetAddress()]; ok {
		return Error.New("connection already exists", nil)
	}
	client.addressConnections[connection.GetAddress()] = connection
	client.nameConnections[connection.GetName()] = connection
	client.waitGroup.Add(1)
	go client.handleDisconnect(connection, reconnectEndpointConfig)
	return nil
}

// AddConnectionAttempt attempts to connect to a server and add it to the client
func (client *SystemgeClient) AddConnectionAttempt(endpointConfig *Config.TcpEndpoint) error {
	if endpointConfig == nil {
		return Error.New("endpointConfig is nil", nil)
	}
	if endpointConfig.Address == "" {
		return Error.New("endpointConfig.Address is empty", nil)
	}
	client.statusMutex.RLock()
	defer client.statusMutex.RUnlock()
	if client.status == Status.STOPPED {
		return Error.New("client stopped", nil)
	}
	return client.startConnectionAttempts(endpointConfig)
}

// RemoveConnection attempts to remove a connection from the client
func (client *SystemgeClient) RemoveConnection(address string) error {
	if address == "" {
		return Error.New("address is empty", nil)
	}
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status == Status.STOPPED {
		return Error.New("client stopped", nil)
	}
	if connection, ok := client.addressConnections[address]; ok {
		connection.Close()
		return nil
	}
	if connectionAttempt, ok := client.connectionAttemptsMap[address]; ok {
		connectionAttempt.isAborted = true
		return nil
	}
	return Error.New("connection not found", nil)
}

func (client *SystemgeClient) GetConnectionByName(name string) SystemgeConnection.SystemgeConnection {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return nil
	}
	return client.nameConnections[name]
}

func (client *SystemgeClient) GetConnectionByAddress(address string) SystemgeConnection.SystemgeConnection {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return nil
	}
	return client.addressConnections[address]
}

func (client *SystemgeClient) GetConnectionNamesAndAddresses() map[string]string {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return nil
	}
	names := make(map[string]string, len(client.addressConnections))
	for address, connection := range client.addressConnections {
		names[connection.GetName()] = address
	}
	return names
}

func (client *SystemgeClient) GetConnectionCount() int {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return 0
	}
	return len(client.addressConnections)
}
