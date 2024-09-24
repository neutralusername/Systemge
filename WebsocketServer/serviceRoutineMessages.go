package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

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

func (server *WebsocketServer) receiveMessage(websocketConnection *WebsocketConnection) error {
	messageBytes, err := server.receive(websocketConnection, Event.ClientReceptionRoutine)
	if err != nil {
		return err
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
	return nil
}
