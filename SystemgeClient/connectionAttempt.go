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
	connectionAttempt := TcpSystemgeConnection.EstablishConnectionAttempts(client.name, &Config.SystemgeConnectionAttempt{
		MaxServerNameLength:         client.config.MaxServerNameLength,
		MaxConnectionAttempts:       client.config.MaxConnectionAttempts,
		RetryIntervalMs:             uint32(client.config.ConnectionAttemptDelayMs),
		EndpointConfig:              endpointConfig,
		TcpSystemgeConnectionConfig: client.config.TcpSystemgeConnectionConfig,
	})
	client.connectionAttemptsMap[endpointConfig.Address] = connectionAttempt
	client.waitGroup.Add(1)

	go func() {
		client.ongoingConnectionAttempts.Add(1)
		client.status = Status.PENDING

		client.handleConnectionAttempt(connectionAttempt)

		val := client.ongoingConnectionAttempts.Add(-1)
		if val == 0 {
			client.status = Status.STARTED
		}
		client.waitGroup.Done()
	}()
	return nil
}

func (client *SystemgeClient) handleConnectionAttempt(connectionAttempt *TcpSystemgeConnection.ConnectionAttempt) {
	endAttempt := func() {
		connectionAttempt.AbortAttempts()
		client.mutex.Lock()
		defer client.mutex.Unlock()
		delete(client.connectionAttemptsMap, connectionAttempt.GetEndpointConfig().Address)
	}

	select {
	case <-client.stopChannel:
		endAttempt()
		return
	case <-connectionAttempt.GetOngoingChannel():
	}

	systemgeConnection, err := connectionAttempt.GetResultBlocking()
	if err != nil {
		if client.errorLogger != nil {
			client.errorLogger.Log(Error.New("Connection attempt failed", err).Error())
		}
		if client.mailer != nil {
			err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("Connection attempt failed", err).Error()))
			if err != nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("failed sending mail", err).Error())
				}
			}
		}
		endAttempt()
		return
	}
	client.connectionAttemptsSuccess.Add(1)

	client.mutex.Lock()
	delete(client.connectionAttemptsMap, connectionAttempt.GetEndpointConfig().Address)
	if client.nameConnections[systemgeConnection.GetName()] != nil {
		client.mutex.Unlock()
		systemgeConnection.Close()
		return
	}
	client.addressConnections[connectionAttempt.GetEndpointConfig().Address] = systemgeConnection
	client.nameConnections[systemgeConnection.GetName()] = systemgeConnection
	client.mutex.Unlock()

	if client.onConnectHandler != nil {
		if err := client.onConnectHandler(systemgeConnection); err != nil {
			if client.warningLogger != nil {
				client.warningLogger.Log(Error.New("onConnectHandler failed for connection \""+systemgeConnection.GetName()+"\"", err).Error())
			}
			systemgeConnection.Close()

			client.mutex.Lock()
			delete(client.addressConnections, systemgeConnection.GetAddress())
			delete(client.nameConnections, systemgeConnection.GetName())
			client.mutex.Unlock()

			return
		}
	}

	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log("Connection established to \"" + connectionAttempt.GetEndpointConfig().Address + "\" with name \"" + systemgeConnection.GetName() + "\" on attempt #" + Helpers.Uint32ToString(connectionAttempt.GetAttemptsCount()))
	}

	client.waitGroup.Add(1)

	if client.config.Reconnect {
		go client.handleDisconnect(systemgeConnection, connectionAttempt.GetEndpointConfig())
	} else {
		go client.handleDisconnect(systemgeConnection, nil)
	}
}

func (client *SystemgeClient) handleDisconnect(connection SystemgeConnection.SystemgeConnection, endpointConfig *Config.TcpClient) {
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

	if client.onDisconnectHandler != nil {
		client.onDisconnectHandler(connection)
	}
	if endpointConfig != nil {
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
