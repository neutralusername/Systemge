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
		"stopped websocketConnections acception",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
		}),
	))
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConn *websocket.Conn) {
	if event := server.onInfo(Event.NewInfo(
		Event.AcceptingClient,
		"accepting websocketConnection",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    Event.WebsocketConnection,
			Event.Address: websocketConn.RemoteAddr().String(),
		}),
	)); event.IsError() {
		if websocketConn != nil {
			websocketConn.Close()
			server.waitGroup.Done()
			server.rejectedWebsocketConnectionsCounter.Add(1)
		}
		return
	}

	server.websocketConnectionMutex.Lock()
	websocketId := server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := server.websocketConnections[websocketId]; exists; {
		websocketId = server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	}
	websocketConnection := &WebsocketConnection{
		id:                  websocketId,
		websocketConnection: websocketConn,
		stopChannel:         make(chan bool),
	}
	if server.config.WebsocketConnectionRateLimiterBytes != nil {
		websocketConnection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(server.config.WebsocketConnectionRateLimiterBytes)
	}
	if server.config.WebsocketConnectionRateLimiterMessages != nil {
		websocketConnection.rateLimiterMsgs = Tools.NewTokenBucketRateLimiter(server.config.WebsocketConnectionRateLimiterMessages)
	}
	websocketConnection.websocketConnection.SetReadLimit(int64(server.config.IncomingMessageByteLimit))
	server.websocketConnections[websocketId] = websocketConnection
	server.websocketConnectionGroups[websocketId] = make(map[string]bool)
	server.websocketConnectionMutex.Unlock()

	go func() {
		select {
		case <-websocketConnection.stopChannel:
		case <-server.stopChannel:
		}

		server.onInfo(Event.NewInfoNoOption(
			Event.DisconnectingClient,
			"disconnecting websocketConnection",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))

		server.removeWebsocketConnection(websocketConnection)
		server.waitGroup.Done()

		server.onInfo(Event.NewInfoNoOption(
			Event.DisconnectedClient,
			"websocketConnection disconnected",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:        Event.WebsocketConnection,
				Event.Address:     websocketConnection.GetIp(),
				Event.WebsocketId: websocketConnection.GetId(),
			}),
		))
	}()

	if event := server.onInfo(Event.NewInfo(
		Event.AcceptedClient,
		"websocketConnection accepted",
		Event.Cancel,
		Event.Continue,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConn.RemoteAddr().String(),
			Event.WebsocketId: websocketId,
		}),
	)); event.IsError() {
		websocketConnection.Close()
		return
	}
	server.acceptedWebsocketConnectionsCounter.Add(1)
	websocketConnection.isAccepted = true
	go server.receiveMessagesLoop(websocketConnection)
}

func (server *WebsocketServer) receiveMessagesLoop(websocketConnection *WebsocketConnection) {
	if event := server.onInfo(Event.NewInfo(
		Event.ReceiveMessageRoutineStarted,
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
		server.incomingMessageCounter.Add(1)
		server.bytesReceivedCounter.Add(uint64(len(messageBytes)))

		if server.config.ExecuteMessageHandlersSequentially {
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
			go func() {
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
		Event.ReceiveMessageRoutineFinished,
		"stopped websocketConnection message reception",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:        Event.WebsocketConnection,
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
		}),
	))
}

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
			Event.FailedToDeserialize,
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
			"no websocketConnection message handler",
			server.GetServerContext().Merge(Event.Context{
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
			Event.Address:     websocketConnection.GetIp(),
			Event.WebsocketId: websocketConnection.GetId(),
			Event.Topic:       message.GetTopic(),
		}),
	))
}
