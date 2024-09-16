package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type ConnectionAttempt struct {
	config             *Config.SystemgeConnectionAttempt
	attempts           uint32
	isAborted          bool
	ongoing            chan bool
	err                error
	systemgeConnection SystemgeConnection.SystemgeConnection
}

func (connectionAttempt *ConnectionAttempt) AbortAttempt() error {
	select {
	case <-connectionAttempt.ongoing:
		return Error.New("Connection attempt is already complete", nil)
	default:
		connectionAttempt.isAborted = true
		return nil
	}
}

func (connectionAttempt *ConnectionAttempt) GetAttempts() uint32 {
	return connectionAttempt.attempts
}

func (connectionAttempt *ConnectionAttempt) BlockUntilComplete() {
	<-connectionAttempt.ongoing
}

func (connectionAttempt *ConnectionAttempt) GetEndpointConfig() *Config.TcpClient {
	return connectionAttempt.config.EndpointConfig
}

func (connectionAttempt *ConnectionAttempt) IsOngoing() bool {
	select {
	case <-connectionAttempt.ongoing:
		return false
	default:
		return true
	}
}

func (connectionAttempt *ConnectionAttempt) GetResult() (SystemgeConnection.SystemgeConnection, error) {
	select {
	case <-connectionAttempt.ongoing:
		return connectionAttempt.systemgeConnection, connectionAttempt.err
	default:
		return nil, Error.New("Connection attempt is ongoing", nil)
	}
}

func StartConnectionAttempts(name string, config *Config.SystemgeConnectionAttempt) (*ConnectionAttempt, error) {
	normalizedAddress, err := Helpers.NormalizeAddress(config.EndpointConfig.Address)
	if err != nil {
		return nil, Error.New("failed normalizing address", err)
	}
	config.EndpointConfig.Address = normalizedAddress

	attempt := &ConnectionAttempt{
		config:   config,
		attempts: 0,
		ongoing:  make(chan bool),
	}

	go connectionAttempts(name, attempt)
	return nil, nil
}

func connectionAttempts(name string, connectionAttempt *ConnectionAttempt) {
	for {
		if connectionAttempt.isAborted {
			connectionAttempt.err = Error.New("Connection attempt aborted before establishing connection", nil)
			close(connectionAttempt.ongoing)
			return
		}
		if connectionAttempt.config.MaxConnectionAttempts > 0 && connectionAttempt.attempts >= connectionAttempt.config.MaxConnectionAttempts {
			connectionAttempt.err = Error.New("Max connection attempts reached", nil)
			close(connectionAttempt.ongoing)
			return
		}
		connectionAttempt.attempts++
		connection, err := EstablishConnection(connectionAttempt.config.TcpSystemgeConnectionConfig, connectionAttempt.config.EndpointConfig, name, connectionAttempt.config.MaxServerNameLength)
		if err != nil {
			continue
		}
		if connectionAttempt.isAborted {
			connection.Close()
			connectionAttempt.err = Error.New("Connection attempt aborted after establishing connection", nil)
			close(connectionAttempt.ongoing)
			return
		}
		connectionAttempt.systemgeConnection = connection
		close(connectionAttempt.ongoing)
		return
	}
}
