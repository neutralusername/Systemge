package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *SystemgeClient) startConnectionAttempts(tcpClientConfig *Config.TcpClient) error {
	normalizedAddress, err := Helpers.NormalizeAddress(tcpClientConfig.Address)
	if err != nil {
		return Event.New("failed normalizing address", err)
	}
	tcpClientConfig.Address = normalizedAddress

	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.addressConnections[tcpClientConfig.Address] != nil {
		return Event.New("Connection already exists", nil)
	}
	if client.connectionAttemptsMap[tcpClientConfig.Address] != nil {
		return Event.New("Connection attempt already in progress", nil)
	}
	connectionAttempt := TcpSystemgeConnection.EstablishConnectionAttempts(client.name, &Config.SystemgeConnectionAttempt{
		MaxServerNameLength:         client.config.MaxServerNameLength,
		MaxConnectionAttempts:       client.config.MaxConnectionAttempts,
		RetryIntervalMs:             uint32(client.config.ConnectionAttemptDelayMs),
		TcpClientConfig:             tcpClientConfig,
		TcpSystemgeConnectionConfig: client.config.TcpSystemgeConnectionConfig,
	})
	client.connectionAttemptsMap[tcpClientConfig.Address] = connectionAttempt
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
		delete(client.connectionAttemptsMap, connectionAttempt.GetTcpClientConfig().Address)
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
			client.errorLogger.Log(Event.New("Connection attempt failed", err).Error())
		}
		if client.mailer != nil {
			err := client.mailer.Send(Tools.NewMail(nil, "error", Event.New("Connection attempt failed", err).Error()))
			if err != nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Event.New("failed sending mail", err).Error())
				}
			}
		}
		endAttempt()
		return
	}
	client.connectionAttemptsSuccess.Add(1)

	client.mutex.Lock()
	delete(client.connectionAttemptsMap, connectionAttempt.GetTcpClientConfig().Address)
	if client.nameConnections[systemgeConnection.GetName()] != nil {
		client.mutex.Unlock()
		systemgeConnection.Close()
		return
	}
	client.addressConnections[connectionAttempt.GetTcpClientConfig().Address] = systemgeConnection
	client.nameConnections[systemgeConnection.GetName()] = systemgeConnection
	client.mutex.Unlock()

	if client.onConnectHandler != nil {
		if err := client.onConnectHandler(systemgeConnection); err != nil {
			if client.warningLogger != nil {
				client.warningLogger.Log(Event.New("onConnectHandler failed for connection \""+systemgeConnection.GetName()+"\"", err).Error())
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
		infoLogger.Log("Connection established to \"" + connectionAttempt.GetTcpClientConfig().Address + "\" with name \"" + systemgeConnection.GetName() + "\" on attempt #" + Helpers.Uint32ToString(connectionAttempt.GetAttemptsCount()))
	}

	client.waitGroup.Add(1)

	if client.config.Reconnect {
		go client.handleDisconnect(systemgeConnection, connectionAttempt.GetTcpClientConfig())
	} else {
		go client.handleDisconnect(systemgeConnection, nil)
	}
}

func (client *SystemgeClient) handleDisconnect(connection SystemgeConnection.SystemgeConnection, tcpClientConfig *Config.TcpClient) {
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
	if tcpClientConfig != nil {
		if err := client.startConnectionAttempts(tcpClientConfig); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Event.New("failed starting (re-)connection attempts to \""+tcpClientConfig.Address+"\"", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Event.New("failed starting (re-)connection attempts to \""+tcpClientConfig.Address+"\"", err).Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Event.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}
	client.waitGroup.Done()
}
