package TcpSystemgeConnect

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type ConnectionAttempt struct {
	config             *Config.SystemgeConnectionAttempt
	attempts           uint32
	isAborted          bool
	ongoing            chan bool
	err                error
	systemgeConnection SystemgeConnection.SystemgeConnection
	eventHandler       Event.Handler
}

func EstablishConnectionAttempts(name string, config *Config.SystemgeConnectionAttempt, eventHandler Event.Handler) (*ConnectionAttempt, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	connectionAttempts := &ConnectionAttempt{
		config:       config,
		attempts:     0,
		ongoing:      make(chan bool),
		eventHandler: eventHandler,
	}
	go connectionAttempts.connectionAttempts(name)
	return connectionAttempts, nil
}

func (connectionAttempt *ConnectionAttempt) AbortAttempts() error {
	select {
	case <-connectionAttempt.ongoing:
		return errors.New("connection attempt has already ended")
	default:
		connectionAttempt.isAborted = true
		return nil
	}
}

func (connectionAttempt *ConnectionAttempt) GetAttemptsCount() uint32 {
	return connectionAttempt.attempts
}

func (connectionAttempt *ConnectionAttempt) GetOngoingChannel() <-chan bool {
	return connectionAttempt.ongoing
}

func (connectionAttempt *ConnectionAttempt) GetTcpClientConfig() *Config.TcpClient {
	return connectionAttempt.config.TcpClientConfig
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

func (connectionAttempt *ConnectionAttempt) connectionAttempts(name string) {
	for {
		if connectionAttempt.isAborted {
			connectionAttempt.err = errors.New("connection attempt aborted before establishing connection")
			close(connectionAttempt.ongoing)
			return
		}
		if connectionAttempt.config.MaxConnectionAttempts > 0 && connectionAttempt.attempts >= connectionAttempt.config.MaxConnectionAttempts {
			connectionAttempt.err = errors.New("max connection attempts reached")
			close(connectionAttempt.ongoing)
			return
		}
		connectionAttempt.attempts++
		connection, err := EstablishConnection(connectionAttempt.config.TcpSystemgeConnectionConfig, connectionAttempt.config.TcpClientConfig, name, connectionAttempt.config.MaxServerNameLength, connectionAttempt.eventHandler)
		if err != nil {
			time.Sleep(time.Duration(connectionAttempt.config.RetryIntervalMs) * time.Millisecond)
			continue
		}
		if connectionAttempt.isAborted {
			connection.Close()
			connectionAttempt.err = errors.New("connection attempt aborted after establishing connection")
			close(connectionAttempt.ongoing)
			return
		}
		connectionAttempt.systemgeConnection = connection
		close(connectionAttempt.ongoing)
		return
	}
}
