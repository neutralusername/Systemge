package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

// AddConnection attempts to add a connection to the client
func (client *SystemgeClient) AddConnection(endpointConfig *Config.TcpEndpoint) error {
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

func (client *SystemgeClient) GetConnection(name string) *SystemgeConnection.SystemgeConnection {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()
	if client.status != Status.STARTED {
		return nil
	}
	for _, connection := range client.addressConnections {
		if connection.GetName() == name {
			return connection
		}
	}
	return nil
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
