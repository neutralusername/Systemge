package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
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

func (connectionAttempt *ConnectionAttempt) GetOngoingChannel() <-chan bool {
	return connectionAttempt.ongoing
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

func (connectionAttempt *ConnectionAttempt) GetResultBlocking() (SystemgeConnection.SystemgeConnection, error) {
	<-connectionAttempt.ongoing
	return connectionAttempt.systemgeConnection, connectionAttempt.err
}

func StartConnectionAttempts(name string, config *Config.SystemgeConnectionAttempt) *ConnectionAttempt {
	connectionAttempts := &ConnectionAttempt{
		config:   config,
		attempts: 0,
		ongoing:  make(chan bool),
	}
	go connectionAttempts.connectionAttempts(name)
	return connectionAttempts
}

func (connectionAttempt *ConnectionAttempt) connectionAttempts(name string) {
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
			time.Sleep(time.Duration(connectionAttempt.config.RetryIntervalMs) * time.Millisecond)
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
