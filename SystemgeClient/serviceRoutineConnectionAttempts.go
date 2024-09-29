package SystemgeClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnect"
)

func (client *SystemgeClient) startConnectionAttempts(tcpClientConfig *Config.TcpClient) error {
	if event := client.onEvent(Event.NewInfo(
		Event.StartingConnectionAttempts,
		"starting connection attempts",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.StartConnectionAttempts,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientAddress: tcpClientConfig.Address,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	normalizedAddress, err := Helpers.NormalizeAddress(tcpClientConfig.Address)
	if err != nil {
		client.onEvent(Event.NewWarningNoOption(
			Event.NormalizingAddressFailed,
			"normalizing address failed",
			Event.Context{
				Event.Circumstance:  Event.StartConnectionAttempts,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientAddress: tcpClientConfig.Address,
			},
		))
		return err
	}
	tcpClientConfig.Address = normalizedAddress

	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.addressConnections[tcpClientConfig.Address] != nil {
		client.onEvent(Event.NewWarningNoOption(
			Event.DuplicateAddress,
			"duplicate address",
			Event.Context{
				Event.Circumstance:  Event.StartConnectionAttempts,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientAddress: tcpClientConfig.Address,
			},
		))
		return errors.New("connection already exists")
	}

	if client.connectionAttemptsMap[tcpClientConfig.Address] != nil {
		client.onEvent(Event.NewWarningNoOption(
			Event.DuplicateAddress,
			"duplicate address",
			Event.Context{
				Event.Circumstance:  Event.StartConnectionAttempts,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientAddress: tcpClientConfig.Address,
				// distinguish this and the previous warning
			},
		))
		return errors.New("connection attempt already in progress")
	}

	connectionAttempt, err := TcpSystemgeConnect.EstablishConnectionAttempts(client.name,
		&Config.SystemgeConnectionAttempt{
			MaxServerNameLength:         client.config.MaxServerNameLength,
			MaxConnectionAttempts:       client.config.MaxConnectionAttempts,
			RetryIntervalMs:             uint32(client.config.ConnectionAttemptDelayMs),
			TcpClientConfig:             tcpClientConfig,
			TcpSystemgeConnectionConfig: client.config.TcpSystemgeConnectionConfig,
		},
		client.eventHandler,
	)
	if err != nil {
		client.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.StartConnectionAttempts,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientAddress: tcpClientConfig.Address,
			},
		))
		return err
	}

	client.connectionAttemptsMap[tcpClientConfig.Address] = connectionAttempt
	client.waitGroup.Add(1)

	go client.handleConnectionAttempt(connectionAttempt)

	client.onEvent(Event.NewInfoNoOption(
		Event.StartedConnectionAttempts,
		"started connection attempts",
		Event.Context{
			Event.Circumstance:  Event.StartConnectionAttempts,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientAddress: tcpClientConfig.Address,
		},
	))
	return nil
}

func (client *SystemgeClient) handleConnectionAttempt(connectionAttempt *TcpSystemgeConnect.ConnectionAttempt) {
	if client.ongoingConnectionAttempts.Add(1) == 1 {
		client.status = Status.Pending
	}
	defer func() {
		if client.ongoingConnectionAttempts.Add(-1) == 0 {
			client.status = Status.Started
		}
		client.waitGroup.Done()
	}()

	endAttempt := func() {
		connectionAttempt.AbortAttempts()
		client.mutex.Lock()
		delete(client.connectionAttemptsMap, connectionAttempt.GetTcpClientConfig().Address)
		client.mutex.Unlock()
		client.connectionAttemptsFailed.Add(uint64(connectionAttempt.GetAttemptsCount()))

		client.onEvent(Event.NewInfoNoOption(
			Event.HandleConnectionAttemptFailed,
			"start connection attempts failed",
			Event.Context{
				Event.Circumstance:  Event.HandleConnectionAttempts,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientAddress: connectionAttempt.GetTcpClientConfig().Address,
			},
		))
	}

	if event := client.onEvent(Event.NewInfo(
		Event.HandlingConnectionAttempt,
		"handling connection attempt",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleConnectionAttempts,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientAddress: connectionAttempt.GetTcpClientConfig().Address,
		},
	)); !event.IsInfo() {
		endAttempt()
		return
	}

	select {
	case <-client.stopChannel:
		endAttempt()
		return
	case <-connectionAttempt.GetOngoingChannel():
	}

	systemgeConnection := connectionAttempt.GetResultBlocking()
	if systemgeConnection == nil {
		endAttempt()
		return
	}
	client.connectionAttemptsFailed.Add(uint64(connectionAttempt.GetAttemptsCount()) - 1)

	err := client.handleAcception(systemgeConnection, connectionAttempt.GetTcpClientConfig())
	if err != nil {
		systemgeConnection.Close()
		client.connectionAttemptsRejected.Add(1)
	}
}

func (client *SystemgeClient) handleAcception(systemgeConnection SystemgeConnection.SystemgeConnection, clientConfig *Config.TcpClient) error {

	if event := client.onEvent(Event.NewInfo(
		Event.HandlingAcception,
		"handling acception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleAcception,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientAddress: clientConfig.Address,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	client.mutex.Lock()
	delete(client.connectionAttemptsMap, clientConfig.Address)

	if _, ok := client.nameConnections[systemgeConnection.GetName()]; !ok {
		client.mutex.Unlock()
		client.onEvent(Event.NewWarningNoOption(
			Event.DuplicateName,
			"duplicate name",
			Event.Context{
				Event.Circumstance:  Event.HandleAcception,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientName:    systemgeConnection.GetName(),
				Event.ClientAddress: clientConfig.Address,
			},
		))
		return errors.New("duplicate name")
	}
	client.addressConnections[clientConfig.Address] = nil
	client.nameConnections[systemgeConnection.GetName()] = nil
	client.mutex.Unlock()

	if event := client.onEvent(Event.NewInfo(
		Event.HandledAcception,
		"handled acception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleAcception,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    systemgeConnection.GetName(),
			Event.ClientAddress: clientConfig.Address,
		},
	)); !event.IsInfo() {
		client.mutex.Lock()
		delete(client.addressConnections, systemgeConnection.GetAddress())
		delete(client.nameConnections, systemgeConnection.GetName())
		client.mutex.Unlock()
		return event.GetError()
	}

	client.mutex.Lock()
	client.addressConnections[systemgeConnection.GetAddress()] = systemgeConnection
	client.nameConnections[systemgeConnection.GetName()] = systemgeConnection
	client.mutex.Unlock()

	client.connectionAttemptsSuccess.Add(1)
	client.waitGroup.Add(1)

	if client.config.AutoReconnectAttempts {
		go client.handleDisconnect(systemgeConnection, clientConfig)
	} else {
		go client.handleDisconnect(systemgeConnection, nil)
	}

	return nil
}

func (client *SystemgeClient) handleDisconnect(connection SystemgeConnection.SystemgeConnection, tcpClientConfig *Config.TcpClient) {
	select {
	case <-connection.GetCloseChannel():
	case <-client.stopChannel:
		connection.Close()
	}

	client.onEvent(Event.NewInfoNoOption(
		Event.HandlingDisconnection,
		"handling disconnect",
		Event.Context{
			Event.Circumstance:  Event.HandleDisconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))

	client.mutex.Lock()
	delete(client.addressConnections, connection.GetAddress())
	delete(client.nameConnections, connection.GetName())
	client.mutex.Unlock()

	if tcpClientConfig != nil {
		if err := client.startConnectionAttempts(tcpClientConfig); err != nil {
			if event := client.onEvent(Event.NewInfo(
				Event.StartConnectionAttemptsFailed,
				"start connection attempts failed",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.HandleDisconnection,
					Event.ClientAddress: tcpClientConfig.Address,
				},
			)); !event.IsInfo() {
				if err := client.stop(false); err != nil {
					panic(err)
				}
				return
			}
		}
	}

	client.onEvent(Event.NewInfoNoOption(
		Event.HandledDisconnection,
		"handled disconnect",
		Event.Context{
			Event.Circumstance:  Event.HandleDisconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))

	client.waitGroup.Done()
}
