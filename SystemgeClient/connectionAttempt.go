package SystemgeClient

import (
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type ConnectionAttempt struct {
	attempts       uint32
	endpointConfig *Config.TcpEndpoint

	isAborted bool
}

func (client *SystemgeClient) startConnectionAttempts(endpointConfig *Config.TcpEndpoint) error {
	noramlzedAddres, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return Error.New("failed normalizing address", err)
	}
	endpointConfig.Address = noramlzedAddres
	client.waitGroup.Add(1)

	client.mutex.Lock()
	if client.connections[endpointConfig.Address] != nil {
		client.mutex.Unlock()
		client.waitGroup.Done()
		return Error.New("Connection already exists", nil)
	}
	if client.connectionAttemptsMap[endpointConfig.Address] != nil {
		client.mutex.Unlock()
		client.waitGroup.Done()
		return Error.New("Connection attempt already in progress", nil)
	}
	attempt := &ConnectionAttempt{
		attempts:       0,
		endpointConfig: endpointConfig,
	}
	client.connectionAttemptsMap[endpointConfig.Address] = attempt
	client.mutex.Unlock()

	go func() {
		defer client.waitGroup.Done()
		val := client.statusUpdateCounter.Add(1)
		if val == 1 {
			client.status = Status.PENDING
		}
		if err := client.connectionAttempts(attempt); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Error.New("failed connection attempt", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed connection attempt", err).Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Error.New("failed sending mail", err).Error())
					}
				}
			}
		}
		val = client.statusUpdateCounter.Add(-1)
		if val == 0 {
			client.status = Status.STARTED
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
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("failed establishing connection attempt #"+Helpers.Uint32ToString(attempt.attempts), err).Error())
				}
				if client.mailer != nil {
					err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed establishing connection attempt", err).Error()))
					if err != nil {
						if client.errorLogger != nil {
							client.errorLogger.Log(Error.New("failed sending mail", err).Error())
						}
					}
				}
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

			client.waitGroup.Add(1)
			go client.connectionClosure(connection, attempt.endpointConfig)
			return nil
		}
	}
}

func (client *SystemgeClient) connectionClosure(connection *SystemgeConnection.SystemgeConnection, endpointConfig *Config.TcpEndpoint) {
	select {
	case <-client.stopChannel:
		connection.Close()
		client.mutex.Lock()
		delete(client.connections, endpointConfig.Address)
		client.mutex.Unlock()
		client.waitGroup.Done()
	case <-connection.GetCloseChannel():
		client.mutex.Lock()
		delete(client.connections, endpointConfig.Address)
		client.mutex.Unlock()
		if client.config.Reconnect {
			if err := client.startConnectionAttempts(endpointConfig); err != nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("failed starting connection attempt", err).Error())
				}
				if client.mailer != nil {
					err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed starting connection attempt", err).Error()))
					if err != nil {
						if client.errorLogger != nil {
							client.errorLogger.Log(Error.New("failed sending mail", err).Error())
						}
					}
				}
			}
		}
		client.waitGroup.Done()
	}
}
