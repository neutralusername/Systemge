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
		}),
	)); event.IsError() {
		return
	}

	for {
		websocketConnection, event := server.receiveWebsocketConnectionFromChannel()
		if event.IsError() {
			break
		}
		if event.IsWarning() {
			continue
		}
		go server.acceptWebsocketConnection(websocketConnection)
	}

	server.onInfo(Event.New(
		Event.AcceptClientsRoutineFinished,
		server.GetServerContext().Merge(Event.Context{
			"info": "websocketServer stopped",
		}),
	))
}

func (server *WebsocketServer) receiveWebsocketConnectionFromChannel() (*websocket.Conn, *Event.Event) {
	if event := server.onInfo(Event.New(
		Event.ReceivingFromChannel,
		server.GetServerContext().Merge(Event.Context{
			"info":        "receiving connection from channel",
			"channelType": "websocketConnection",
		}),
	)); event.IsError() {
		return nil, event
	}
	websocketConnection := <-server.connectionChannel
	if websocketConnection == nil {
		return nil, server.onError(Event.New(
			Event.ReceivedNilValueFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"error":       "received nil value from channel",
				"channelType": "websocketConnection",
			}),
		))
	}
	return websocketConnection, server.onInfo(Event.New(
		Event.ReceivedFromChannel,
		server.GetServerContext().Merge(Event.Context{
			"info":        "received connection from channel",
			"address":     websocketConnection.RemoteAddr().String(),
			"channelType": "websocketConnection",
		}),
	))
}

func (server *WebsocketServer) acceptWebsocketConnection(websocketConnection *websocket.Conn) {
	if event := server.onInfo(Event.New(
		Event.AcceptingClient,
		server.GetServerContext().Merge(Event.Context{
			"info":    "accepting websocket connection",
			"address": websocketConnection.RemoteAddr().String(),
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
			"address":     websocketConnection.RemoteAddr().String(),
			"websocketId": websocketId,
		}),
	)); event.IsError() {
		return
	}
	client.isAccepted = true
	go server.receiveMessagesLoop(client)
}

func (server *WebsocketServer) receiveMessagesLoop(client *WebsocketClient) {
	if event := server.onInfo(Event.New(
		Event.ReceiveMessageRoutineStarted,
		server.GetServerContext().Merge(Event.Context{
			"info":        "started receiving messages from client",
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

		event = server.handleClientMessage(client, messageBytes)
		if event.IsError() {
			break
		}
	}

	server.onInfo(Event.New(
		Event.ReceiveMessageRoutineFinished,
		server.GetServerContext().Merge(Event.Context{
			"info":        "stopped receiving messages from client",
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
				"warning":         "rate limited",
				"rateLimiterType": "bytes",
				"address":         client.GetIp(),
				"websocketId":     client.GetId(),
			}),
		))
	}
	if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
		return server.onWarning(Event.New(
			Event.RateLimited,
			server.GetServerContext().Merge(Event.Context{
				"warning":         "rate limited",
				"rateLimiterType": "messages",
				"address":         client.GetIp(),
				"websocketId":     client.GetId(),
			}),
		))
	}
	message, err := Message.Deserialize(messageBytes, client.GetId())
	if err != nil {
		return server.onError(Event.New(
			Event.FailedToDeserialize,
			server.GetServerContext().Merge(Event.Context{
				"error":       "failed to deserialize message",
				"address":     client.GetIp(),
				"websocketId": client.GetId(),
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

	return server.onInfo(Event.New(
		Event.HandledMessage,
		server.GetServerContext().Merge(Event.Context{
			"info":               "handled message from client",
			"serviceRoutineType": "handleMessages",
			"address":            client.GetIp(),
			"websocketId":        client.GetId(),
		}),
	))
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
