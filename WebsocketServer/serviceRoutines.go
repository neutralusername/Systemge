package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
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
	} else {
		server.acceptWebsocketConnection(websocketConnection)
	}
	return nil
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

	for {
		messageBytes, err := server.receive(websocketConnection, Event.ClientReceptionRoutine)
		if err != nil {
			break
		}
		server.websocketConnectionMessagesReceived.Add(1)
		server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

		if server.config.HandleMessagesSequentially {
			event := server.handleWebsocketConnectionMessage(websocketConnection, messageBytes)
			if event.IsError() {
				if server.config.PropagateMessageHandlerErrors {
					server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ClientReceptionRoutine)
				}
				websocketConnection.Close()
			}
			if event.IsWarning() {
				if server.config.PropagateMessageHandlerWarnings {
					server.send(websocketConnection, Message.NewAsync("warning", event.Marshal()).Serialize(), Event.ClientReceptionRoutine)
				}
			}
		} else {
			websocketConnection.waitGroup.Add(1)
			go func() {
				defer websocketConnection.waitGroup.Done()
				event := server.handleWebsocketConnectionMessage(websocketConnection, messageBytes)
				if event.IsError() {
					if server.config.PropagateMessageHandlerErrors {
						server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ClientReceptionRoutine)
					}
					websocketConnection.Close()
				}
				if event.IsWarning() {
					if server.config.PropagateMessageHandlerWarnings {
						server.send(websocketConnection, Message.NewAsync("warning", event.Marshal()).Serialize(), Event.ClientReceptionRoutine)
					}
				}
			}()
		}
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
