package connectionTcp

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
)

type ConnectionAttempt struct {
	config             *configs.SystemgeConnectionAttempt
	systemgeConnection systemge.Connection[[]byte]

	attempts   uint32
	ongoing    chan bool
	abortMutex sync.Mutex
}

func EstablishConnectionAttempts(name string, config *configs.SystemgeConnectionAttempt) (*ConnectionAttempt, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	connectionAttempts := &ConnectionAttempt{
		config:   config,
		attempts: 0,
		ongoing:  make(chan bool),
	}
	go connectionAttempts.connectionAttempts(name)
	return connectionAttempts, nil
}
func (connectionAttempt *ConnectionAttempt) connectionAttempts(name string) {
	for {
		select {
		case <-connectionAttempt.ongoing:
			return
		default:
		}
		if connectionAttempt.config.MaxConnectionAttempts > 0 && connectionAttempt.attempts >= connectionAttempt.config.MaxConnectionAttempts {
			connectionAttempt.AbortAttempts()
			return
		}
		connectionAttempt.attempts++
		connection, err := EstablishConnection(connectionAttempt.config.TcpBufferedReaderConfig, connectionAttempt.config.TcpClientConfig)
		if err != nil {
			select {
			case <-time.After(time.Duration(connectionAttempt.config.RetryIntervalMs) * time.Millisecond):
			case <-connectionAttempt.ongoing:
				return
			}
			continue
		}
		select {
		case <-connectionAttempt.ongoing:
			connection.Close()
			return
		default:
		}
		connectionAttempt.systemgeConnection = connection
		connectionAttempt.AbortAttempts()
		return
	}
}

func (connectionAttempt *ConnectionAttempt) AbortAttempts() error {
	connectionAttempt.abortMutex.Lock()
	defer connectionAttempt.abortMutex.Unlock()
	select {
	case <-connectionAttempt.ongoing:
		return errors.New("connection attempt has already ended")
	default:
		close(connectionAttempt.ongoing)
		return nil
	}
}

func (connectionAttempt *ConnectionAttempt) GetAttemptsCount() uint32 {
	return connectionAttempt.attempts
}

func (connectionAttempt *ConnectionAttempt) GetOngoingChannel() <-chan bool {
	return connectionAttempt.ongoing
}

func (connectionAttempt *ConnectionAttempt) GetTcpClientConfig() *configs.TcpClient {
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

func (connectionAttempt *ConnectionAttempt) GetResultBlocking() systemge.Connection[[]byte] {
	<-connectionAttempt.ongoing
	return connectionAttempt.systemgeConnection
}
