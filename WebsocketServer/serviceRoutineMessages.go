package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) receptionRoutine(websocketConnection *WebsocketConnection) {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.ReceptionRoutineFinished,
			"reception routine finished",
			Event.Context{
				Event.Circumstance:  Event.ReceptionRoutine,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
			},
		))
		websocketConnection.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.ReceptionRoutineStarted,
		"reception routine started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.ReceptionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.receiveMessage(websocketConnection); err == nil; {
	}

}

func (server *WebsocketServer) receiveMessage(websocketConnection *WebsocketConnection) error {
	messageBytes, err := server.receive(websocketConnection, Event.ReceptionRoutine)
	if err != nil {
		return err
	}
	server.websocketConnectionMessagesReceived.Add(1)
	server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

	if server.config.HandleMessageReceptionSequentially {
		event := server.handleReception(websocketConnection, messageBytes)
		if event.IsError() {
			if server.config.PropagateMessageHandlerErrors {
				server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
			}
			websocketConnection.Close()
		}
		if event.IsWarning() {
			if server.config.PropagateMessageHandlerWarnings {
				server.send(websocketConnection, Message.NewAsync("warning", event.Marshal()).Serialize(), Event.ReceptionRoutine)
			}
		}
	} else {
		websocketConnection.waitGroup.Add(1)
		go func() {
			defer websocketConnection.waitGroup.Done()
			event := server.handleReception(websocketConnection, messageBytes)
			if event.IsError() {
				if server.config.PropagateMessageHandlerErrors {
					server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
				}
				websocketConnection.Close()
			}
			if event.IsWarning() {
				if server.config.PropagateMessageHandlerWarnings {
					server.send(websocketConnection, Message.NewAsync("warning", event.Marshal()).Serialize(), Event.ReceptionRoutine)
				}
			}
		}()
	}
	return nil
}

func (server *WebsocketServer) handleReception(websocketConnection *WebsocketConnection, messageBytes []byte) *Event.Event {
	event := server.onEvent(Event.NewInfo(
		Event.HandlingReception,
		"handling websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		},
	))
	if !event.IsInfo() {
		return event
	}

	if websocketConnection.rateLimiterBytes != nil && !websocketConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := server.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleReception,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Bytes,
				Event.ClientType:      Event.WebsocketConnection,
				Event.ClientId:        websocketConnection.GetId(),
				Event.ClientAddress:   websocketConnection.GetIp(),
			},
		)); !event.IsInfo() {
			return event
		}
	}

	if websocketConnection.rateLimiterMsgs != nil && !websocketConnection.rateLimiterMsgs.Consume(1) {
		if event := server.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection message rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleReception,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
				Event.ClientType:      Event.WebsocketConnection,
				Event.ClientId:        websocketConnection.GetId(),
				Event.ClientAddress:   websocketConnection.GetIp(),
			},
		)); !event.IsInfo() {
			return event
		}
	}

	message, err := Message.Deserialize(messageBytes, websocketConnection.GetId())
	if err != nil {
		return server.onEvent(Event.NewWarningNoOption(
			Event.DeserializingFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.StructType:    Event.Message,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
				Event.Bytes:         string(messageBytes),
			},
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload()) // getting rid of possible syncToken
	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		return server.onEvent(Event.NewInfoNoOption(
			Event.HeartbeatReceived,
			"received websocketConnection heartbeat",
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
			},
		))
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		return server.onEvent(Event.NewWarningNoOption(
			Event.NoHandlerForTopic,
			"no websocketConnection message handler for topic",
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.HandlerType:   Event.WebsocketConnection,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
				Event.Topic:         message.GetTopic(),
			},
		))
	}

	if err := handler(websocketConnection, message); err != nil {
		return server.onEvent(Event.NewWarningNoOption(
			Event.HandlerFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.HandlerType:   Event.WebsocketConnection,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
				Event.Topic:         message.GetTopic(),
			},
		))
	}

	return server.onEvent(Event.NewInfoNoOption(
		Event.HandledReception,
		"handled websocketConnection message",
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.HandlerType:   Event.WebsocketConnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
			Event.Topic:         message.GetTopic(),
		},
	))
}
