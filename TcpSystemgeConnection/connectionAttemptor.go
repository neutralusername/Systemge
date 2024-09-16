package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type ConnectionAttemptor struct {
	config             *Config.SystemgeConnectionAttemptor
	attempts           uint32
	isAborted          bool
	ongoing            chan bool
	err                error
	systemgeConnection SystemgeConnection.SystemgeConnection
}

func (connectionAttempt *ConnectionAttemptor) AbortAttempt() error {
	select {
	case <-connectionAttempt.ongoing:
		return Error.New("Connection attempt is already complete", nil)
	default:
		connectionAttempt.isAborted = true
		return nil
	}
}

func (connectionAttempt *ConnectionAttemptor) GetAttempts() uint32 {
	return connectionAttempt.attempts
}

func (connectionAttempt *ConnectionAttemptor) BlockUntilComplete() {
	<-connectionAttempt.ongoing
}

func (connectionAttempt *ConnectionAttemptor) GetEndpointConfig() *Config.TcpClient {
	return connectionAttempt.config.EndpointConfig
}

func (connectionAttempt *ConnectionAttemptor) IsOngoing() bool {
	select {
	case <-connectionAttempt.ongoing:
		return false
	default:
		return true
	}
}

func (connectionAttempt *ConnectionAttemptor) GetResult() (SystemgeConnection.SystemgeConnection, error) {
	select {
	case <-connectionAttempt.ongoing:
		return connectionAttempt.systemgeConnection, connectionAttempt.err
	default:
		return nil, Error.New("Connection attempt is ongoing", nil)
	}
}

func StartConnectionAttempts(name string, config *Config.SystemgeConnectionAttemptor) *ConnectionAttemptor {
	connectionAttempts := &ConnectionAttemptor{
		config:   config,
		attempts: 0,
		ongoing:  make(chan bool),
	}
	go connectionAttempts.connectionAttempts(name)
	return connectionAttempts
}

func (connectionAttempt *ConnectionAttemptor) connectionAttempts(name string) {
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
