package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

// AddConnection adds an active connection to the client.
// if reconnectTcpClientConfig is not nil, the connection will attempt to reconnect
func (client *SystemgeClient) AddConnection(connection SystemgeConnection.SystemgeConnection, reconnectTcpClientConfig *Config.TcpClient) error {
	if connection == nil {
		return Event.New("connection is nil", nil)
	}
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status == Status.STOPPED {
		return Event.New("client stopped", nil)
	}
	if _, ok := client.addressConnections[connection.GetAddress()]; ok {
		return Event.New("connection already exists", nil)
	}
	client.addressConnections[connection.GetAddress()] = connection
	client.nameConnections[connection.GetName()] = connection
	client.waitGroup.Add(1)
	go client.handleDisconnect(connection, reconnectTcpClientConfig)
	return nil
}

// AddConnectionAttempt attempts to connect to a server and add it to the client
func (client *SystemgeClient) AddConnectionAttempt(tcpClientConfig *Config.TcpClient) error {
	if tcpClientConfig == nil {
		return Event.New("tcpClientConfig is nil", nil)
	}
	if tcpClientConfig.Address == "" {
		return Event.New("tcpClientConfig.Address is empty", nil)
	}
	client.statusMutex.RLock()
	defer client.statusMutex.RUnlock()
	if client.status == Status.STOPPED {
		return Event.New("client stopped", nil)
	}
	return client.startConnectionAttempts(tcpClientConfig)
}

// RemoveConnection attempts to remove a connection from the client
func (client *SystemgeClient) RemoveConnection(address string) error {
	if address == "" {
		return Event.New("address is empty", nil)
	}
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status == Status.STOPPED {
		return Event.New("client stopped", nil)
	}
	if connection, ok := client.addressConnections[address]; ok {
		connection.Close()
		return nil
	}
	if connectionAttempt, ok := client.connectionAttemptsMap[address]; ok {
		return connectionAttempt.AbortAttempts()
	}
	return Event.New("connection not found", nil)
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

func (client *SystemgeClient) GetConnectionName(address string) string {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return ""
	}
	connection, ok := client.addressConnections[address]
	if !ok {
		return ""
	}
	return connection.GetName()
}

func (client *SystemgeClient) GetConnectionAddress(name string) string {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return ""
	}
	connection, ok := client.nameConnections[name]
	if !ok {
		return ""
	}
	return connection.GetAddress()
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
