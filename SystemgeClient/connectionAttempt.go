package SystemgeClient

import (
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type ConnectionAttempt struct {
	attempts       uint32
	endpointConfig *Config.TcpEndpoint

	isAborted bool
}

func (client *SystemgeClient) startConnectionAttempts(endpointConfig *Config.TcpEndpoint) error {
	client.connectionAttemptWaitGroup.Add(1)
	client.startingConnectionAttemptChannel <- true

	client.mutex.Lock()
	if client.connections[endpointConfig.Address] != nil {
		client.mutex.Unlock()
		return Error.New("Connection already exists", nil)
	}
	if client.connectionAttemptsMap[endpointConfig.Address] != nil {
		client.mutex.Unlock()
		return Error.New("Connection attempt already in progress", nil)
	}
	attempt := &ConnectionAttempt{
		attempts:       0,
		endpointConfig: endpointConfig,
	}
	client.connectionAttemptsMap[endpointConfig.Address] = attempt
	client.mutex.Unlock()

	go func() {
		defer client.connectionAttemptWaitGroup.Done()
		err := client.connectionAttempts(attempt)
		if err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Error.New("failed connection attempt", err).Error())
			}
		}
	}()
	return nil
}

func (client *SystemgeClient) connectionAttempts(attempt *ConnectionAttempt) error {
	endAttempt := func() {
		client.mutex.Lock()
		defer client.mutex.Unlock()
		delete(client.connectionAttemptsMap, attempt.endpointConfig.Address)
	}
	for {
		select {
		case <-client.stopChannel:
			endAttempt()
			return Error.New("SystemgeClient stopped", nil)
		default:
			if attempt.isAborted {
				endAttempt()
				return Error.New("Connection attempt aborted", nil)
			}
			if client.config.MaxConnectionAttempts > 0 && attempt.attempts >= client.config.MaxConnectionAttempts {
				endAttempt()
				return Error.New("Max connection attempts reached", nil)
			}
			if attempt.attempts > 0 {
				time.Sleep(time.Duration(client.config.ConnectionAttemptDelayMs) * time.Millisecond)
			}
			connection, err := SystemgeConnection.EstablishConnection(client.config.ConnectionConfig, attempt.endpointConfig, client.GetName(), client.config.MaxServerNameLength, client.messageHandler)
			if err != nil {
				client.connectionAttemptsFailed.Add(1)
				attempt.attempts++
				continue
			}
			client.connectionAttemptsSuccess.Add(1)

			client.mutex.Lock()
			defer client.mutex.Unlock()
			delete(client.connectionAttemptsMap, attempt.endpointConfig.Address)
			if attempt.isAborted {
				connection.Close()
				return Error.New("Connection attempt aborted", nil)
			}
			client.connections[attempt.endpointConfig.Address] = connection

			client.connectionWaitGroup.Add(1)
			go func() {
				select {
				case <-client.stopChannel:
					connection.Close()
					client.mutex.Lock()
					delete(client.connections, attempt.endpointConfig.Address)
					client.mutex.Unlock()
					client.connectionWaitGroup.Done()
				case <-connection.GetCloseChannel():
					if client.config.Reconnect {
						client.mutex.Lock()
						delete(client.connections, attempt.endpointConfig.Address)
						client.mutex.Unlock()
						go func() {
							if err := client.startConnectionAttempts(attempt.endpointConfig); err != nil {
								if client.errorLogger != nil {
									client.errorLogger.Log(Error.New("failed connection attempt", err).Error())
								}
							}
						}()
					}
					client.connectionWaitGroup.Done()
				}
			}()
			return nil
		}
	}
}
