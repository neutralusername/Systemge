package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

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
