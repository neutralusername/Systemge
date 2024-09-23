package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) receiveWebsocketConnectionLoop() {
	defer server.waitGroup.Done()

	if event := server.onInfo(Event.NewInfo(
		Event.ClientAcceptionRoutineStarted,
		"started websocketConnections acception",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
		}),
	)); event.IsError() {
		return
	}

	for {
		if event := server.onInfo(Event.NewInfo(
			Event.ReceivingFromChannel,
			"receiving websocketConnection from channel",
			Event.Cancel,
			Event.Continue,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: Event.WebsocketConnection,
			}),
		)); event.IsError() {
			break
		}

		websocketConnection := <-server.connectionChannel
		if websocketConnection == nil {
			server.onInfo(Event.NewInfoNoOption(
				Event.ReceivedNilValueFromChannel,
				"received nil from websocketConnection channel",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: Event.WebsocketConnection,
				}),
			))
			break
		}
		event := server.onInfo(Event.NewInfo(
			Event.ReceivedFromChannel,
			"received websocketConnection from channel",
			Event.Cancel,
			Event.Skip,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    Event.WebsocketConnection,
				Event.Address: websocketConnection.RemoteAddr().String(),
			}),
		))
		if event.IsError() {
			websocketConnection.Close()
			server.websocketConnectionsRejected.Add(1)
			break
		}
		if event.IsWarning() {
			websocketConnection.Close()
			server.websocketConnectionsRejected.Add(1)
			continue
		}

		server.acceptWebsocketConnection(websocketConnection)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientAcceptionRoutineFinished,
		"stopped websocketConnections acception",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
		}),
	))
}

func (server *WebsocketServer) receiveMessagesLoop(websocketConnection *WebsocketConnection) {
	defer websocketConnection.waitGroup.Done()

	if event := server.onInfo(Event.NewInfo(
		Event.ClientMessageReceptionRoutineStarted,
		"started websocketConnection message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
		}),
	)); !event.IsInfo() {
		return
	}

	for {
		messageBytes, err := server.receive(websocketConnection)
		if err != nil {
			break
		}
		server.websocketConnectionMessagesReceived.Add(1)
		server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

		if server.config.HandleMessagesSequentially {
			event := server.handleWebsocketConnectionMessage(websocketConnection, messageBytes)
			if event.IsError() {
				if server.config.PropagateMessageHandlerErrors {
					server.Send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize())
				}
				break
			}
			if event.IsWarning() {
				if server.config.PropagateMessageHandlerWarnings {
					server.Send(websocketConnection, Message.NewAsync("warning", event.Marshal()).Serialize())
				}
			}
		} else {
			websocketConnection.waitGroup.Add(1)
			go func() {
				defer websocketConnection.waitGroup.Done()
				event := server.handleWebsocketConnectionMessage(websocketConnection, messageBytes)
				if event.IsError() {
					if server.config.PropagateMessageHandlerErrors {
						server.Send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize())
					}
				}
				if event.IsWarning() {
					if server.config.PropagateMessageHandlerWarnings {
						server.Send(websocketConnection, Message.NewAsync("warning", event.Marshal()).Serialize())
					}
				}
			}()
		}
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientMessageReceptionRoutineFinished,
		"stopped websocketConnection message reception",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
		}),
	))
}
