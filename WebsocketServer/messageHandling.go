package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) handleWebsocketConnectionMessage(websocketConnection *WebsocketConnection, messageBytes []byte) *Event.Event {
	event := server.onEvent(Event.NewInfo(
		Event.HandlingClientMessage,
		"handling websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.MessageHandlingRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
		}),
	))
	if !event.IsInfo() {
		return event
	}

	if websocketConnection.rateLimiterBytes != nil && !websocketConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := server.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection message byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.MessageHandlingRoutine,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Bytes,
				Event.ClientType:      Event.WebsocketConnection,
				Event.ClientId:        websocketConnection.GetId(),
				Event.ClientAddress:   websocketConnection.GetIp(),
			}),
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
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.MessageHandlingRoutine,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
				Event.ClientType:      Event.WebsocketConnection,
				Event.ClientId:        websocketConnection.GetId(),
				Event.ClientAddress:   websocketConnection.GetIp(),
			}),
		)); !event.IsInfo() {
			return event
		}
	}

	message, err := Message.Deserialize(messageBytes, websocketConnection.GetId())
	if err != nil {
		return server.onEvent(Event.NewWarningNoOption(
			Event.DeserializingFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.MessageHandlingRoutine,
				Event.StructType:    Event.Message,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
			}),
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload()) // getting rid of possible syncToken
	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		return server.onEvent(Event.NewInfoNoOption(
			Event.HeartbeatReceived,
			"received websocketConnection heartbeat",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.MessageHandlingRoutine,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
			}),
		))
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		return server.onEvent(Event.NewWarningNoOption(
			Event.NoHandlerForTopic,
			"no websocketConnection message handler for topic",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.MessageHandlingRoutine,
				Event.HandlerType:   Event.WebsocketConnection,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
				Event.Topic:         message.GetTopic(),
			}),
		))
	}

	if err := handler(websocketConnection, message); err != nil {
		return server.onEvent(Event.NewWarningNoOption(
			Event.HandlerFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:  Event.MessageHandlingRoutine,
				Event.HandlerType:   Event.WebsocketConnection,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetIp(),
				Event.Topic:         message.GetTopic(),
			}),
		))
	}

	return server.onEvent(Event.NewInfoNoOption(
		Event.HandledClientMessage,
		"handled websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:  Event.MessageHandlingRoutine,
			Event.HandlerType:   Event.WebsocketConnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetIp(),
			Event.Topic:         message.GetTopic(),
		}),
	))
}
