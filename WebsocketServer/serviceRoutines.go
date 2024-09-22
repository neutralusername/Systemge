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
		server.GetServerContext().Merge(Event.Context{
			"info": "websocketServer started",
			"type": "websocketConnection",
		}),
	)); event.IsError() {
		return
	}

	for {
		if event := server.onInfo(Event.New(
			Event.ReceivingFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info": "receiving connection from channel",
				"type": "websocketConnection",
			}),
		)); event.IsError() {
			break
		}
		websocketConnection := <-server.connectionChannel
		if websocketConnection == nil {
			server.onInfo(Event.New(
				Event.ReceivedNilValueFromChannel,
				server.GetServerContext().Merge(Event.Context{
					"info": "received nil value from channel",
					"type": "websocketConnection",
				}),
			))
			break
		}
		event := server.onInfo(Event.New(
			Event.ReceivedFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":    "received connection from channel",
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
		server.GetServerContext().Merge(Event.Context{
			"info": "websocketServer stopped",
			"type": "websocketConnection",
		}),
	))
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConnection *websocket.Conn) {
	if event := server.onInfo(Event.New(
		Event.AcceptingClient,
		server.GetServerContext().Merge(Event.Context{
			"info":    "accepting websocket connection",
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
			server.GetServerContext().Merge(Event.Context{
				"info":        "websocket connection disconnected",
				"type":        "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))

		server.removeClient(client)
		server.waitGroup.Done()

		server.onInfo(Event.New(
			Event.DisconnectedClient,
			server.GetServerContext().Merge(Event.Context{
				"info":        "websocket connection disconnected",
				"type":        "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))

	}()

	if event := server.onInfo(Event.New(
		Event.AcceptedClient,
		server.GetServerContext().Merge(Event.Context{
			"info":        "websocket connection accepted",
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
		server.GetServerContext().Merge(Event.Context{
			"info":        "started receiving messages from client",
			"type":        "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	)); event.IsError() {
		return
	}

	for {
		messageBytes, event := server.receive(client)
		if event.IsError() {
			break
		}
		if event.IsWarning() {
			continue
		}
		server.incomingMessageCounter.Add(1)
		server.bytesReceivedCounter.Add(uint64(len(messageBytes)))

		if server.config.ExecuteMessageHandlersSequentially {
			event = server.handleClientMessage(client, messageBytes)
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
		server.GetServerContext().Merge(Event.Context{
			"info":        "stopped receiving messages from client",
			"type":        "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
}

func (server *WebsocketServer) handleClientMessage(client *WebsocketClient, messageBytes []byte) *Event.Event {
	event := server.onInfo(Event.New(
		Event.HandlingMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":        "handling message from client",
			"type":        "websocketConnection",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
		}),
	))
	if event.IsError() {
		return event
	}

	if client.rateLimiterBytes != nil && !client.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		return server.onWarning(Event.New(
			Event.RateLimited,
			server.GetServerContext().Merge(Event.Context{
				"warning":     "bytes rate limited",
				"type":        "tokenBucket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}

	if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
		return server.onWarning(Event.New(
			Event.RateLimited,
			server.GetServerContext().Merge(Event.Context{
				"warning":     "messages rate limited",
				"type":        "tokenBucket",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}

	message, err := Message.Deserialize(messageBytes, client.GetId())
	if err != nil {
		return server.onError(Event.New(
			Event.FailedToDeserialize,
			server.GetServerContext().Merge(Event.Context{
				"error":       err.Error(),
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload()) // getting rid of possible syncToken
	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		return server.onInfo(Event.New(
			Event.HeartbeatReceived,
			server.GetServerContext().Merge(Event.Context{
				"info":        "received heartbeat from client",
				"type":        "websocketConnection",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
			}),
		))
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		return server.onWarning(Event.New(
			Event.NoHandlerForTopic,
			server.GetServerContext().Merge(Event.Context{
				"warning":     "no handler for for provided topic",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
				"topic":       message.GetTopic(),
			}),
		))
	}

	if err := handler(client, message); err != nil {
		return server.onWarning(Event.New(
			Event.HandlerFailed,
			server.GetServerContext().Merge(Event.Context{
				"warning":     err.Error(),
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
				"topic":       message.GetTopic(),
			}),
		))
	}

	return server.onInfo(Event.New(
		Event.HandledMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":        "handled message from client",
			"address":     client.GetIp(),
			"websocketId": client.GetId(),
			"topic":       message.GetTopic(),
		}),
	))
}
