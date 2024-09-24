package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
)

func (server *WebsocketServer) receiveWebsocketConnectionLoop() {
	defer server.waitGroup.Done()

	if event := server.onEvent(Event.NewInfo(
		Event.ClientAcceptionRoutineStarted,
		"started websocketConnections acception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.ClientAcceptionRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		}),
	)); !event.IsInfo() {
		return
	}

	for err := server.receiveWebsocketConnection(); err == nil; {
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.ClientAcceptionRoutineFinished,
		"stopped websocketConnections acception",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.ClientAcceptionRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		}),
	))
}

func (server *WebsocketServer) receiveWebsocketConnection() error {
	if event := server.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"receiving websocketConnection from channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.ClientAcceptionRoutine,
			Event.ChannelType:  Event.WebsocketConnection,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	websocketConnection := <-server.connectionChannel
	if websocketConnection == nil {
		server.onEvent(Event.NewInfoNoOption(
			Event.ReceivedNilValueFromChannel,
			"received nil from websocketConnection channel",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.ClientAcceptionRoutine,
				Event.ChannelType:  Event.WebsocketConnection,
			}),
		))
		return errors.New("received nil from websocketConnection channel")
	}

	event := server.onEvent(Event.NewInfo(
		Event.ReceivedFromChannel,
		"received websocketConnection from channel",
		Event.Cancel,
		Event.Skip,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientAcceptionRoutine,
			Event.ChannelType:   Event.WebsocketConnection,
			Event.ClientAddress: websocketConnection.RemoteAddr().String(),
		}),
	))
	if event.IsError() {
		websocketConnection.Close()
		server.websocketConnectionsRejected.Add(1)
		return event.GetError()
	}
	if event.IsWarning() {
		websocketConnection.Close()
		server.websocketConnectionsRejected.Add(1)
		return nil
	}

	server.acceptWebsocketConnection(websocketConnection)
	return nil
}
