package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type ConnectionAttempt struct {
	attempts       uint32
	endpointConfig *Config.TcpClient

	isAborted bool
}

func (client *SystemgeClient) startConnectionAttempts(endpointConfig *Config.TcpClient) error {
	normalizedAddress, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return Error.New("failed normalizing address", err)
	}
	endpointConfig.Address = normalizedAddress

	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.addressConnections[endpointConfig.Address] != nil {
		return Error.New("Connection already exists", nil)
	}
	if client.connectionAttemptsMap[endpointConfig.Address] != nil {
		return Error.New("Connection attempt already in progress", nil)
	}
	attempt := &ConnectionAttempt{
		attempts:       0,
		endpointConfig: endpointConfig,
	}
	client.connectionAttemptsMap[endpointConfig.Address] = attempt

	client.waitGroup.Add(1)

	go func() {
		client.ongoingConnectionAttempts.Add(1)
		client.status = Status.PENDING
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
		val := client.ongoingConnectionAttempts.Add(-1)
		if val == 0 {
			client.status = Status.STARTED
		}
		client.waitGroup.Done()
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
			connection, err := TcpSystemgeConnection.EstablishConnection(client.config.TcpSystemgeConnectionConfig, attempt.endpointConfig, client.GetName(), client.config.MaxServerNameLength)
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
			delete(client.connectionAttemptsMap, attempt.endpointConfig.Address)
			if attempt.isAborted {
				client.mutex.Unlock()
				connection.Close()
				return Error.New("Connection attempt aborted", nil)
			}
			if client.nameConnections[connection.GetName()] != nil {
				client.mutex.Unlock()
				connection.Close()
				return Error.New("Connection name already exists", nil)
			}
			client.addressConnections[attempt.endpointConfig.Address] = connection
			client.nameConnections[connection.GetName()] = connection
			client.mutex.Unlock()

			if client.onConnectHandler != nil {
				if err := client.onConnectHandler(connection); err != nil {
					if client.warningLogger != nil {
						client.warningLogger.Log(Error.New("onConnectHandler failed for connection \""+connection.GetName()+"\"", err).Error())
					}
					connection.Close()

					client.mutex.Lock()
					delete(client.addressConnections, connection.GetAddress())
					delete(client.nameConnections, connection.GetName())
					client.mutex.Unlock()

					return Error.New("onConnectHandler failed", err)
				}
			}

			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log("Connection established to \"" + attempt.endpointConfig.Address + "\" with name \"" + connection.GetName() + "\" on attempt #" + Helpers.Uint32ToString(attempt.attempts))
			}

			client.waitGroup.Add(1)
			go client.handleDisconnect(connection)
			return nil
		}
	}
}

func (client *SystemgeClient) handleDisconnect(connection SystemgeConnection.SystemgeConnection) {
	select {
	case <-connection.GetCloseChannel():
	case <-client.stopChannel:
		connection.Close()
	}

	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log("Connection closed to \"" + connection.GetAddress() + "\" with name \"" + connection.GetName() + "\"")
	}

	client.mutex.Lock()
	delete(client.addressConnections, connection.GetAddress())
	delete(client.nameConnections, connection.GetName())
	client.mutex.Unlock()

	var reconnectEndpointConfig *Config.TcpClient
	if client.onDisconnectHandler != nil {
		reconnectEndpointConfig = client.onDisconnectHandler(connection)
	}

	if reconnectEndpointConfig != nil {
		if err := client.startConnectionAttempts(reconnectEndpointConfig); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Error.New("failed starting (re-)connection attempts to \""+reconnectEndpointConfig.Address+"\"", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed starting (re-)connection attempts to \""+reconnectEndpointConfig.Address+"\"", err).Error()))
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
