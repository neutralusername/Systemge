package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (server *WebsocketServer) handleWebsocketConnections() {
	if event := server.onInfo(Event.New(
		Event.ServiceRoutineStarted,
		server.GetServerContext().Merge(Event.Context{
			"info":               "websocketServer started",
			"serviceRoutineType": "acceptClients",
		}),
	)); event.IsError() {
		return
	}

	for {
		if event := server.onInfo(Event.New(
			Event.ReceivingFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":               "receiving connection from channel",
				"serviceRoutineType": "acceptClients",
				"channelType":        "websocketConnection",
			}),
		)); event.IsError() {
			break
		}
		websocketConnection := <-server.connectionChannel
		if websocketConnection == nil {
			server.onError(Event.New(
				Event.ReceivedNilValueFromChannel,
				server.GetServerContext().Merge(Event.Context{
					"error":              "received nil value from channel",
					"serviceRoutineType": "acceptClients",
					"channelType":        "websocketConnection",
				}),
			))

		}
		event := server.onInfo(Event.New(
			Event.ReceivedFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":               "received connection from channel",
				"serviceRoutineType": "acceptClients",
				"address":            websocketConnection.RemoteAddr().String(),
				"channelType":        "websocketConnection",
			}),
		))
		if event.IsError() {
			break
		}
		if event.IsWarning() {
			continue
		}
		go server.handleWebsocketConnection(websocketConnection)
	}

	server.onInfo(Event.New(
		Event.ServiceRoutineFinished,
		server.GetServerContext().Merge(Event.Context{
			"info":               "websocketServer stopped",
			"serviceRoutineType": "acceptClients",
		}),
	))
}

func (server *WebsocketServer) handleWebsocketConnection(websocketConnection *websocket.Conn) {
	if event := server.onInfo(Event.New(
		Event.AcceptingClient,
		server.GetServerContext().Merge(Event.Context{
			"info":       "accepting websocket connection",
			"acceptType": "websocketClient",
			"address":    websocketConnection.RemoteAddr().String(),
		}),
	)); event.IsError() {
		return
	}

	server.clientMutex.Lock()
	websocketId := server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := server.clients[websocketId]; exists; {
		websocketId = server.randomizer.GenerateRandomString(16, Tools.ALPHA_NUMERIC)
	}
	client := server.newClient(websocketId, websocketConnection)
	server.clients[websocketId] = client
	server.clientGroups[websocketId] = make(map[string]bool)
	server.clientMutex.Unlock()

	defer client.Disconnect()
	if event := server.onInfo(Event.New( // websocketClient can be acquired in the handler function through its id
		Event.AcceptedClient,
		server.GetServerContext().Merge(Event.Context{
			"info":        "websocket connection accepted",
			"acceptType":  "websocketClient",
			"address":     websocketConnection.RemoteAddr().String(),
			"websocketId": websocketId,
		}),
	)); event.IsError() {
		return
	}
	client.isAccepted = true
	go server.handleMessages(client)
}

func (server *WebsocketServer) handleMessages(client *WebsocketClient) {
	if event := server.onInfo(Event.New(
		Event.ServiceRoutineStarted,
		server.GetServerContext().Merge(Event.Context{
			"info":               "handling messages from client",
			"serviceRoutineType": "handleMessages",
			"address":            client.GetIp(),
			"websocketId":        client.GetId(),
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
		event = server.handleClientMessage(client, messageBytes)
		if event.IsError() {
			break
		}
		if event.IsWarning() {
			continue
		}
	}
	if event := server.onInfo(Event.New(
		Event.ServiceRoutineFinished,
		server.GetServerContext().Merge(Event.Context{
			"info":               "handling messages from client finished",
			"serviceRoutineType": "handleMessages",
			"address":            client.GetIp(),
			"websocketId":        client.GetId(),
		}),
	)); event.IsError() {
		return
	}
}

func (server *WebsocketServer) handleClientMessage(client *WebsocketClient, messageBytes []byte) *Event.Event {
	if client.rateLimiterBytes != nil && !client.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		return server.onWarning(Event.New(
			Event.FailedToHandleMessage,
			server.GetServerContext().Merge(Event.Context{
				"warning":            "rate limited",
				"serviceRoutineType": "handleMessages",
				"address":            client.GetIp(),
				"websocketId":        client.GetId(),
			}),
		))
	}
	if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
		return server.onWarning(Event.New(
			Event.FailedToHandleMessage,
			server.GetServerContext().Merge(Event.Context{
				"warning":            "rate limited",
				"serviceRoutineType": "handleMessages",
				"address":            client.GetIp(),
				"websocketId":        client.GetId(),
			}),
		))
	}
	message, err := Message.Deserialize(messageBytes, client.GetId())
	if err != nil {
		return server.onError(Event.New(
			Event.FailedToHandleMessage,
			server.GetServerContext().Merge(Event.Context{
				"error":              "failed to deserialize message",
				"serviceRoutineType": "handleMessages",
				"address":            client.GetIp(),
				"websocketId":        client.GetId(),
			}),
		))
	}
	message = Message.NewAsync(message.GetTopic(), message.GetPayload())
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Event.New("Received message with topic \""+message.GetTopic()+"\" from client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", nil).Error())
	}
	if message.GetTopic() == "heartbeat" {
		server.ResetWatchdog(client)
		continue
	}
	if server.config.HandleClientMessagesSequentially {
		err := server.handleClientMessage(client, message)
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Event.New("Failed to handle message (sequentially)", err).Error())
			}
		} else {
			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Event.New("Handled message (sequentially)", nil).Error())
			}
		}
	} else {
		go func() {
			err := server.handleClientMessage(client, message)
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Event.New("Failed to handle message (concurrently)", err).Error())
				}
			} else {
				if infoLogger := server.infoLogger; infoLogger != nil {
					infoLogger.Log(Event.New("Handled message (concurrently)", nil).Error())
				}
			}
		}()
	}
}

func (server *WebsocketServer) handleClientMessage(client *WebsocketClient, message *Message.Message) error {
	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()
	if handler == nil {
		err := client.Send(Message.NewAsync("error", Event.New("no handler for topic \""+message.GetTopic()+"\" from client \""+client.GetId()+"\"", nil).Error()).Serialize())
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Event.New("Failed to send error message to client", err).Error())
			}
		}
		return Event.New("no handler for topic \""+message.GetTopic()+"\"", nil)
	}
	if err := handler(client, message); err != nil {
		if err := client.Send(Message.NewAsync("error", Event.New("error in handler for topic \""+message.GetTopic()+"\" from client \""+client.GetId()+"\"", err).Error()).Serialize()); err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Event.New("Failed to send error message to client", err).Error())
			}
		}
	}
	return nil
}
