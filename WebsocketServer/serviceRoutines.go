package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (server *WebsocketServer) receiveWebsocketConnectionLoop() {
	if event := server.onInfo(Event.New(
		Event.AcceptClientsRoutineStarted,
		"websocketServer started",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type": "websocketConnection",
		}),
	)); event.IsError() {
		return
	}

	for {
		if event := server.onInfo(Event.New(
			Event.ReceivingFromChannel,
			"receiving connection from channel",
			Event.Info,
			Event.Cancel,
			Event.Continue,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				"type": "websocketConnection",
			}),
		)); event.IsError() {
			break
		}
		websocketConnection := <-server.connectionChannel
		if websocketConnection == nil {
			server.onInfo(Event.New(
				Event.ReceivedNilValueFromChannel,
				"received nil value from channel",
				Event.Info,
				Event.NoOption,
				Event.NoOption,
				Event.NoOption,
				server.GetServerContext().Merge(Event.Context{
					"type": "websocketConnection",
				}),
			))
			break
		}
		event := server.onInfo(Event.New(
			Event.ReceivedFromChannel,
			"received connection from channel",
			Event.Info,
			Event.Cancel,
			Event.Skip,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				"type":    "websocketConnection",
				"address": websocketConnection.RemoteAddr().String(),
			}),
		))
		if event.IsError() {
			websocketConnection.Close()
			server.waitGroup.Done()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			break
		}
		if event.IsWarning() {
			websocketConnection.Close()
			server.waitGroup.Done()
			server.rejectedWebsocketConnectionsCounter.Add(1)
			continue
		}
		go server.acceptWebsocketConnection(websocketConnection)
	}

	server.onInfo(Event.New(
		Event.AcceptClientsRoutineFinished,
		"websocketServer stopped",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			"type": "websocketConnection",
		}),
	))
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConnection *websocket.Conn) {
	if event := server.onInfo(Event.New(
		Event.AcceptingClient,
		"accepting websocket connection",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":    "websocketConnection",
			"address": websocketConnection.RemoteAddr().String(),
		}),
	)); event.IsError() {
		if websocketConnection != nil {
			websocketConnection.Close()
			server.waitGroup.Done()
			server.rejectedWebsocketConnectionsCounter.Add(1)
		}
		return
	}

	server.clientMutex.Lock()
	websocketId := server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := server.clients[websocketId]; exists; {
		websocketId = server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	}
	client := &WebsocketClient{
		id:                  websocketId,
		websocketConnection: websocketConnection,
		stopChannel:         make(chan bool),
	}
	if server.config.ClientRateLimiterBytes != nil {
		client.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.ClientRateLimiterBytes)
	}
	if server.config.ClientRateLimiterMessages != nil {
		client.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.ClientRateLimiterMessages)
	}
	client.websocketConnection.SetReadLimit(int64(server.config.IncomingMessageByteLimit))
	server.clients[websocketId] = client
	server.clientGroups[websocketId] = make(map[string]bool)
	server.clientMutex.Unlock()

	go func() {
		select {
		case <-client.stopChannel:
		case <-server.stopChannel:
		}

		server.onInfo(Event.New(
			Event.DisconnectingClient,
			"disconnecting websocket connection",
			Event.Info,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				"type":        "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))

		server.removeClient(client)
		server.waitGroup.Done()

		server.onInfo(Event.New(
			Event.DisconnectedClient,
			"websocket connection disconnected",
			Event.Info,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				"type":        "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}()

	if event := server.onInfo(Event.New(
		Event.AcceptedClient,
		"websocket connection accepted",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":        "websocketConnection",
			"address":     websocketConnection.RemoteAddr().String(),
			"websocketId": websocketId,
		}),
	)); event.IsError() {
		client.Close()
		return
	}
	server.acceptedWebsocketConnectionsCounter.Add(1)
	client.isAccepted = true
	go server.receiveMessagesLoop(client)
}

func (server *WebsocketServer) receiveMessagesLoop(client *WebsocketClient) {
	if event := server.onInfo(Event.New(
		Event.ReceiveMessageRoutineStarted,
		"started receiving messages from client",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":        "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	)); event.IsError() {
		return
	}

	for {
		messageBytes, err := server.receive(client)
		if err != nil {
			break
		}
		server.incomingMessageCounter.Add(1)
		server.bytesReceivedCounter.Add(uint64(len(messageBytes)))

		if server.config.ExecuteMessageHandlersSequentially {
			event := server.handleClientMessage(client, messageBytes)
			if event.IsError() {
				if server.config.PropagateMessageHandlerErrors {
					server.Send(client, Message.NewAsync("error", event.Marshal()).Serialize())
				}
				break
			}
			if event.IsWarning() {
				if server.config.PropagateMessageHandlerWarnings {
					server.Send(client, Message.NewAsync("warning", event.Marshal()).Serialize())
				}
			}
		} else {
			go func() {
				event := server.handleClientMessage(client, messageBytes)
				if event.IsError() {
					if server.config.PropagateMessageHandlerErrors {
						server.Send(client, Message.NewAsync("error", event.Marshal()).Serialize())
					}
				}
				if event.IsWarning() {
					if server.config.PropagateMessageHandlerWarnings {
						server.Send(client, Message.NewAsync("warning", event.Marshal()).Serialize())
					}
				}
			}()
		}
	}

	server.onInfo(Event.New(
		Event.ReceiveMessageRoutineFinished,
		"stopped receiving messages from client",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			"type":        "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
}

func (server *WebsocketServer) handleClientMessage(client *WebsocketClient, messageBytes []byte) *Event.Event {
	event := server.onInfo(Event.New(
		Event.HandlingMessage,
		"handling message from client",
		Event.Info,
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			"type":        "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
	if event.IsError() {
		return event
	}

	if client.rateLimiterBytes != nil && !client.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := server.onError(Event.New(
			Event.RateLimited,
			"byte rate limited",
			Event.Error,
			Event.Cancel,
			Event.Continue,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				"type":        "tokenBucket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		)); event.IsError() {
			return event
		}
	}

	if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
		if server.onError(Event.New(
			Event.RateLimited,
			"message rate limited",
			Event.Error,
			Event.Cancel,
			Event.Continue,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				"type":        "tokenBucket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		)); event.IsError() {
			return event
		}
	}

	message, err := Message.Deserialize(messageBytes, client.GetId())
	if err != nil {
		return server.onError(Event.New(
			Event.FailedToDeserialize,
			err.Error(),
			Event.Error,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload()) // getting rid of possible syncToken
	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		return server.onInfo(Event.New(
			Event.HeartbeatReceived,
			"received heartbeat from client",
			Event.Info,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				"type": "websocketConnection",
			}),
		))
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		return server.onError(Event.New(
			Event.NoHandlerForTopic,
			"no handler for provided topic",
			Event.Warning,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
				"topic":       message.GetTopic(),
			}),
		))
	}

	if err := handler(client, message); err != nil {
		return server.onError(Event.New(
			Event.HandlerFailed,
			err.Error(),
			Event.Warning,
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
			server.GetServerContext().Merge(Event.Context{
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
				"topic":       message.GetTopic(),
			}),
		))
	}

	return server.onInfo(Event.New(
		Event.HandledMessage,
		"handled message from client",
		Event.Info,
		Event.NoOption,
		Event.NoOption,
		Event.NoOption,
		server.GetServerContext().Merge(Event.Context{
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
			"topic":       message.GetTopic(),
		}),
	))
}
