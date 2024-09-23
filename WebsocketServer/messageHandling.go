package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) handleWebsocketConnectionMessage(websocketConnection *WebsocketConnection, messageBytes []byte) *Event.Event {
	event := server.onInfo(Event.NewInfo(
		Event.HandlingMessage,
		"handling websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
		}),
	))
	if !event.IsInfo() {
		return event
	}

	if websocketConnection.rateLimiterBytes != nil && !websocketConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := server.onWarning(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection message byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:           Event.TokenBucket,
				Event.AdditionalKind: Event.Bytes,
				Event.Address:        websocketConnection.GetIp(),
				Event.WebsocketId:    websocketConnection.GetId(),
			}),
		)); !event.IsInfo() {
			return event
		}
	}

	if websocketConnection.rateLimiterMsgs != nil && !websocketConnection.rateLimiterMsgs.Consume(1) {
		if event := server.onWarning(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection message rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:           Event.TokenBucket,
				Event.AdditionalKind: Event.Messages,
				Event.Address:        websocketConnection.GetIp(),
				Event.WebsocketId:    websocketConnection.GetId(),
			}),
		)); !event.IsInfo() {
			return event
		}
	}

	message, err := Message.Deserialize(messageBytes, websocketConnection.GetId())
	if err != nil {
		return server.onWarning(Event.NewWarningNoOption(
			Event.DeserializingMessageFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload()) // getting rid of possible syncToken
	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		return server.onInfo(Event.NewInfoNoOption(
			Event.HeartbeatReceived,
			"received websocketConnection heartbeat",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: Event.WebsocketConnection,
			}),
		))
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		return server.onWarning(Event.NewWarningNoOption(
			Event.NoHandlerForTopic,
			"no websocketConnection message handler for topic",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
				Event.Topic:       message.GetTopic(),
			}),
		))
	}

	if err := handler(websocketConnection, message); err != nil {
		return server.onWarning(Event.NewWarningNoOption(
			Event.HandlerFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
				Event.Topic:       message.GetTopic(),
			}),
		))
	}

	return server.onInfo(Event.NewInfoNoOption(
		Event.HandledMessage,
		"handled websocketConnection message",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
			Event.Topic:       message.GetTopic(),
		}),
	))
}
