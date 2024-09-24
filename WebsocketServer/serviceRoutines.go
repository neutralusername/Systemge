package WebsocketServer

import (
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

func (server *WebsocketServer) receiveMessagesLoop(websocketConnection *WebsocketConnection) {
	defer websocketConnection.waitGroup.Done()

	if event := server.onEvent(Event.NewInfo(
		Event.ClientReceptionRoutineStarted,
		"started websocketConnection message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientReceptionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		}),
	)); !event.IsInfo() {
		return
	}

	for err := server.receiveMessage(websocketConnection); err == nil; {
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.ClientReceptionRoutineFinished,
		"stopped websocketConnection message reception",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.ClientReceptionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		}),
	))
}
