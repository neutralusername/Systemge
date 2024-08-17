package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

// AddConnection attempts to add a connection to the client
func (client *SystemgeClient) AddConnection(endpointConfig *Config.TcpEndpoint) error {
	if endpointConfig == nil {
		return Error.New("endpointConfig is nil", nil)
	}
	if endpointConfig.Address == "" {
		return Error.New("endpointConfig.Address is empty", nil)
	}
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status == Status.STOPPED {
		return Error.New("client stopped", nil)
	}
	return client.startConnectionAttempts(endpointConfig)
}
