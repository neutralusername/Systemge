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
	normalizedAddress, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return Error.New("failed normalizing address", err)
	}
	endpointConfig.Address = normalizedAddress
	client.waitGroup.Add(1)

	client.mutex.Lock()
	if client.addressConnections[endpointConfig.Address] != nil {
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
		val := client.ongoingConnectionAttempts.Add(1)
		if val == 1 {
			client.status = Status.PENDING
		}
		if err := client.connectionAttempts(attempt); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Error.New("failed connection attempts to \""+attempt.endpointConfig.Address+"\"", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed connection attempts to \""+attempt.endpointConfig.Address+"\"", err).Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Error.New("failed sending mail", err).Error())
					}
				}
			}
		}
		val = client.ongoingConnectionAttempts.Add(-1)
		if val == 0 {
			client.status = Status.STARTED
		}
	}()
	return nil
}

func (client *SystemgeClient) connectionAttempts(attempt *ConnectionAttempt) error {
	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log("Starting connection attempts to \"" + attempt.endpointConfig.Address + "\"")
	}
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
			attempt.attempts++
			if err != nil {
				client.connectionAttemptsFailed.Add(1)
				if client.warningLogger != nil {
					client.warningLogger.Log(Error.New("failed establishing connection to \""+attempt.endpointConfig.Address+"\" on attempt #"+Helpers.Uint32ToString(attempt.attempts), err).Error())
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
			if client.nameConnections[connection.GetName()] != nil {
				connection.Close()
				return Error.New("Connection name already exists", nil)
			}
			client.addressConnections[attempt.endpointConfig.Address] = connection
			client.nameConnections[connection.GetName()] = connection

			client.waitGroup.Add(1)

			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log("Connection established to \"" + attempt.endpointConfig.Address + "\" with name \"" + connection.GetName() + "\" on attempt #" + Helpers.Uint32ToString(attempt.attempts))
			}
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
		delete(client.addressConnections, endpointConfig.Address)
		delete(client.nameConnections, connection.GetName())
		client.mutex.Unlock()
		client.waitGroup.Done()
	case <-connection.GetCloseChannel():
		client.mutex.Lock()
		delete(client.addressConnections, endpointConfig.Address)
		delete(client.nameConnections, connection.GetName())
		client.mutex.Unlock()
		if client.config.Reconnect {
			if err := client.startConnectionAttempts(endpointConfig); err != nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("failed starting (re-)connection attempts to \""+endpointConfig.Address+"\"", err).Error())
				}
				if client.mailer != nil {
					err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed starting (re-)connection attempts to \""+endpointConfig.Address+"\"", err).Error()))
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
