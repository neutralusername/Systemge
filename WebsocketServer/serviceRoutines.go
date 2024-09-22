package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (server *WebsocketServer) receiveWebsocketConnectionLoop() {
	if event := server.onInfo(Event.NewInfo(
		Event.AcceptClientsRoutineStarted,
		"websocketServer started",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: "websocketConnection",
		}),
	)); event.IsError() {
		return
	}

	for {
		if event := server.onInfo(Event.NewInfo(
			Event.ReceivingFromChannel,
			"receiving connection from channel",
			Event.Cancel,
			Event.Continue,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: "websocketConnection",
			}),
		)); event.IsError() {
			break
		}
		websocketConnection := <-server.connectionChannel
		if websocketConnection == nil {
			server.onInfo(Event.NewInfoNoOption(
				Event.ReceivedNilValueFromChannel,
				"received nil value from channel",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind: "websocketConnection",
				}),
			))
			break
		}
		event := server.onInfo(Event.NewInfo(
			Event.ReceivedFromChannel,
			"received connection from channel",
			Event.Cancel,
			Event.Skip,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: "websocketConnection",
				"address":  websocketConnection.RemoteAddr().String(),
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

	server.onInfo(Event.NewInfoNoOption(
		Event.AcceptClientsRoutineFinished,
		"websocketServer stopped",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: "websocketConnection",
		}),
	))
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConnection *websocket.Conn) {
	if event := server.onInfo(Event.NewInfo(
		Event.AcceptingClient,
		"accepting websocket connection",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: "websocketConnection",
			"address":  websocketConnection.RemoteAddr().String(),
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

		server.onInfo(Event.NewInfoNoOption(
			Event.DisconnectingClient,
			"disconnecting websocket connection",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))

		server.removeClient(client)
		server.waitGroup.Done()

		server.onInfo(Event.NewInfoNoOption(
			Event.DisconnectedClient,
			"websocket connection disconnected",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}()

	if event := server.onInfo(Event.NewInfo(
		Event.AcceptedClient,
		"websocket connection accepted",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    "websocketConnection",
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
	if event := server.onInfo(Event.NewInfo(
		Event.ReceiveMessageRoutineStarted,
		"started receiving messages from client",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	)); !event.IsInfo() {
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

	server.onInfo(Event.NewInfoNoOption(
		Event.ReceiveMessageRoutineFinished,
		"stopped receiving messages from client",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
}

func (server *WebsocketServer) handleClientMessage(client *WebsocketClient, messageBytes []byte) *Event.Event {
	event := server.onInfo(Event.NewInfo(
		Event.HandlingMessage,
		"handling message from client",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
	if !event.IsInfo() {
		return event
	}

	if client.rateLimiterBytes != nil && !client.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := server.onWarning(Event.NewWarning(
			Event.RateLimited,
			"byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    "tokenBucket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		)); !event.IsInfo() {
			return event
		}
	}

	if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
		if event := server.onWarning(Event.NewWarning(
			Event.RateLimited,
			"message rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    "tokenBucket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		)); !event.IsInfo() {
			return event
		}
	}

	message, err := Message.Deserialize(messageBytes, client.GetId())
	if err != nil {
		return server.onWarning(Event.NewWarningNoOption(
			Event.FailedToDeserialize,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload()) // getting rid of possible syncToken
	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		return server.onInfo(Event.NewInfoNoOption(
			Event.HeartbeatReceived,
			"received heartbeat from client",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: "websocketConnection",
			}),
		))
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		return server.onWarning(Event.NewWarningNoOption(
			Event.NoHandlerForTopic,
			"no handler for provided topic",
			server.GetServerContext().Merge(Event.Context{
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
				"topic":       message.GetTopic(),
			}),
		))
	}

	if err := handler(client, message); err != nil {
		return server.onWarning(Event.NewWarningNoOption(
			Event.HandlerFailed,
			err.Error(),
			server.GetServerContext().Merge(Event.Context{
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
				"topic":       message.GetTopic(),
			}),
		))
	}

	return server.onInfo(Event.NewInfoNoOption(
		Event.HandledMessage,
		"handled message from client",
		server.GetServerContext().Merge(Event.Context{
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
			"topic":       message.GetTopic(),
		}),
	))
}
