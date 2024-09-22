package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (server *WebsocketServer) receiveWebsocketConnectionLoop(stopChannel chan bool) {
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
		websocketConnection, event := server.receiveWebsocketConnectionFromChannel()
		if event.IsError() {
			if websocketConnection != nil {
				websocketConnection.Close()
				server.waitGroup.Done()
				server.rejectedWebsocketConnectionsCounter.Add(1)
			}
			break
		}
		if event.IsWarning() {
			if websocketConnection != nil {
				websocketConnection.Close()
				server.waitGroup.Done()
				server.rejectedWebsocketConnectionsCounter.Add(1)
			}
			continue
		}
		go server.acceptWebsocketConnection(websocketConnection, stopChannel)
	}

	server.onInfo(Event.New(
		Event.AcceptClientsRoutineFinished,
		server.GetServerContext().Merge(Event.Context{
			"info": "websocketServer stopped",
			"type": "websocketConnection",
		}),
	))
}

func (server *WebsocketServer) receiveWebsocketConnectionFromChannel() (*websocket.Conn, *Event.Event) {
	if event := server.onInfo(Event.New(
		Event.ReceivingFromChannel,
		server.GetServerContext().Merge(Event.Context{
			"info": "receiving connection from channel",
			"type": "websocketConnection",
		}),
	)); event.IsError() {
		return nil, event
	}
	select { // receive until all n remaining httpHandlers sent their connections then go to default and wait for stopChannel (if stopChannel triggers, that means no more connections will be received)
	case websocketConnection := <-server.connectionChannel:
		return websocketConnection, server.onInfo(Event.New(
			Event.ReceivedFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":    "received connection from channel",
				"type":    "websocketConnection",
				"address": websocketConnection.RemoteAddr().String(),
			}),
		))
	default:
		select {
		case <-server.stopChannel:
			return nil, server.onError(Event.New(
				Event.ServiceAlreadyStopped,
				server.GetServerContext().Merge(Event.Context{
					"error": "websocketServer stopped",
				}),
			))
		case websocketConnection := <-server.connectionChannel:
			return websocketConnection, server.onInfo(Event.New(
				Event.ReceivedFromChannel,
				server.GetServerContext().Merge(Event.Context{
					"info":    "received connection from channel",
					"type":    "websocketConnection",
					"address": websocketConnection.RemoteAddr().String(),
				}),
			))
		}
	}
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConnection *websocket.Conn, stopChannel chan bool) {
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
		case <-stopChannel:
		}

		server.removeClient(client)

		if server.onDisconnectHandler != nil {
			server.onDisconnectHandler(client)
		}

		server.waitGroup.Done()
	}()

	if event := server.onInfo(Event.New( // websocketClient can be acquired in the handler function through its id
		Event.AcceptedClient,
		server.GetServerContext().Merge(Event.Context{
			"info":        "websocket connection accepted",
			"type":        "websocketConnection",
			"address":     websocketConnection.RemoteAddr().String(),
			"websocketId": websocketId,
		}),
	)); event.IsError() {
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
					bytes, _ := event.Marshal()
					server.Send(client, Message.NewAsync("error", string(bytes)).Serialize())
				}
				break
			}
			if event.IsWarning() {
				if server.config.PropagateMessageHandlerWarnings {
					bytes, _ := event.Marshal()
					server.Send(client, Message.NewAsync("warning", string(bytes)).Serialize())
				}
			}
		} else {
			go func() {
				event := server.handleClientMessage(client, messageBytes)
				if event.IsError() {
					if server.config.PropagateMessageHandlerErrors {
						bytes, _ := event.Marshal()
						server.Send(client, Message.NewAsync("error", string(bytes)).Serialize())
					}
				}
				if event.IsWarning() {
					if server.config.PropagateMessageHandlerWarnings {
						bytes, _ := event.Marshal()
						server.Send(client, Message.NewAsync("warning", string(bytes)).Serialize())
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
