package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
)

func (server *SystemgeServer) acceptRoutine(stopChannel chan bool) {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.AcceptionRoutineFinished,
			"stopped systemgeServer acception routine",
			Event.Context{
				Event.Circumstance: Event.AcceptionRoutine,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.AcceptionRoutineStarted,
		"started systemgeServer acception routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptionRoutine,
			Event.ClientType:   Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.acceptSystemgeConnection(); err == nil; {
	}
}

func (server *SystemgeServer) acceptSystemgeConnection() error {
	select {
	case <-server.stopChannel:
		return errors.New("systemgeServer stopped")
	default:
		if event := server.onEvent(Event.NewInfo(
			Event.AcceptingClient,
			"accepting systemgeConnection",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.AcceptionRoutine,
				Event.ClientType:   Event.SystemgeConnection,
			},
		)); !event.IsInfo() {
			return event.GetError()
		}

		server.waitGroup.Add(1)
		connection, err := server.listener.AcceptConnection(server.config.TcpSystemgeConnectionConfig, server.eventHandler)
		if err != nil {
			event := server.onEvent(Event.NewInfo(
				Event.AcceptingClientFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.AcceptionRoutine,
					Event.ClientType:   Event.SystemgeConnection,
				},
			))
			server.waitGroup.Done()
			if !event.IsInfo() {
				return event.GetError()
			} else {
				return nil
			}
		}

		server.mutex.Lock()
		if _, ok := server.clients[connection.GetName()]; ok {
			event := server.onEvent(Event.NewInfo(
				Event.DuplicateName,
				"duplicate name",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.AcceptionRoutine,
					Event.ClientType:   Event.SystemgeConnection,
					Event.ClientName:   connection.GetName(),
				},
			))
			server.mutex.Unlock()
			connection.Close()
			server.waitGroup.Done()
			if !event.IsInfo() {
				return errors.New("duplicate name")
			} else {
				return nil
			}
		}
		server.clients[connection.GetName()] = connection
		server.mutex.Unlock()

		isAccepted := false
		go func() {
			select {
			case <-connection.GetCloseChannel():
			case <-stopChannel:
				connection.Close()
			}
			server.mutex.Lock()
			delete(server.clients, connection.GetName())
			server.mutex.Unlock()

			if isAccepted {
				if server.onDisconnectHandler != nil {
					server.onDisconnectHandler(connection)
				}
			}

			server.waitGroup.Done()

			if server.infoLogger != nil {
				server.infoLogger.Log("connection \"" + connection.GetName() + "\" closed")
			}
		}()

		if server.onConnectHandler != nil {
			if err := server.onConnectHandler(connection); err != nil {
				connection.Close()
				return
			}
		}
		isAccepted = true

		if server.infoLogger != nil {
			server.infoLogger.Log("connection \"" + connection.GetName() + "\" accepted")
		}
	}
}
