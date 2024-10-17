package Client

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/status"
)

// AddConnection adds an active connection to the client.
// if reconnectTcpClientConfig is not nil, the connection will attempt to reconnect
func (client *Client) AddConnection(connection SystemgeConnection.SystemgeConnection, reconnectTcpClientConfig *Config.TcpClient) error {
	if connection == nil {
		return errors.New("connection is nil")
	}
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status == status.Stopped {
		return errors.New("client stopped")
	}
	if _, ok := client.addressConnections[connection.GetAddress()]; ok {
		return errors.New("connection already exists")
	}
	client.waitGroup.Add(1)
	client.addressConnections[connection.GetAddress()] = connection
	client.nameConnections[connection.GetName()] = connection
	go client.handleDisconnect(connection, reconnectTcpClientConfig)
	return nil
}

// AddConnectionAttempt attempts to connect to a server and add it to the client
func (client *Client) AddConnectionAttempt(tcpClientConfig *Config.TcpClient) error {
	if tcpClientConfig == nil {
		return errors.New("tcpClientConfig is nil")
	}
	if tcpClientConfig.Address == "" {
		return errors.New("tcpClientConfig.Address is empty")
	}
	client.statusMutex.RLock()
	defer client.statusMutex.RUnlock()
	if client.status == status.Stopped {
		return errors.New("client stopped")
	}
	return client.startConnectionAttempts(tcpClientConfig)
}

// RemoveConnection attempts to remove a connection from the client
func (client *Client) RemoveConnection(address string) error {
	if address == "" {
		return errors.New("address is empty")
	}
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status == status.Stopped {
		return errors.New("client stopped")
	}
	if connection, ok := client.addressConnections[address]; ok {
		connection.Close()
		return nil
	}
	if connectionAttempt, ok := client.connectionAttemptsMap[address]; ok {
		return connectionAttempt.AbortAttempts()
	}
	return errors.New("connection not found")
}

func (client *Client) GetConnectionByName(name string) SystemgeConnection.SystemgeConnection {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != status.Started {
		return nil
	}
	return client.nameConnections[name]
}

func (client *Client) GetConnectionByAddress(address string) SystemgeConnection.SystemgeConnection {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != status.Started {
		return nil
	}
	return client.addressConnections[address]
}

func (client *Client) GetConnectionNamesAndAddresses() map[string]string {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != status.Started {
		return nil
	}
	names := make(map[string]string, len(client.addressConnections))
	for address, connection := range client.addressConnections {
		names[connection.GetName()] = address
	}
	return names
}

func (client *Client) GetConnectionName(address string) string {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != status.Started {
		return ""
	}
	connection, ok := client.addressConnections[address]
	if !ok {
		return ""
	}
	return connection.GetName()
}

func (client *Client) GetConnectionAddress(name string) string {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != status.Started {
		return ""
	}
	connection, ok := client.nameConnections[name]
	if !ok {
		return ""
	}
	return connection.GetAddress()
}

func (client *Client) GetConnectionCount() int {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != status.Started {
		return 0
	}
	return len(client.addressConnections)
}
