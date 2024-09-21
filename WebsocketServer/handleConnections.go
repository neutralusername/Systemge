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
			"serviceRoutineType": "handleWebsocketConnections",
		}),
	)); event.IsError() {
		return
	}
	for {
		if event := server.onInfo(Event.New(
			Event.ReceivingFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":               "receiving connection from channel",
				"serviceRoutineType": "handleWebsocketConnections",
			}),
		)); event.IsError() {
			return
		}
		websocketConnection := <-server.connectionChannel
		if websocketConnection == nil {
			server.onInfo(Event.New(
				Event.ServiceRoutineFinished,
				server.GetServerContext().Merge(Event.Context{
					"info":               "websocketServer stopped",
					"serviceRoutineType": "handleWebsocketConnections",
				}),
			))
			return
		}
		if event := server.onInfo(Event.New(
			Event.ReceivedFromChannel,
			server.GetServerContext().Merge(Event.Context{
				"info":               "received connection from channel",
				"serviceRoutineType": "handleWebsocketConnections",
				"address":            websocketConnection.RemoteAddr().String(),
			}),
		)); event.IsError() {
			continue
		}
		go server.handleWebsocketConnection(websocketConnection)
	}
}

func (server *WebsocketServer) handleWebsocketConnection(websocketConnection *websocket.Conn) {
	if event := server.onInfo(Event.New(
		Event.AcceptingClient,
		server.GetServerContext().Merge(Event.Context{
			"info":    "accepting connection",
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
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Event.New("client connected with id \""+client.GetId()+"\" and ip \""+client.GetIp()+"\"", nil).Error())
	}
	if server.onConnectHandler != nil {
		err := server.onConnectHandler(client)
		if err != nil {
			return
		}
	}
	client.pastOnConnectHandler = true
	server.handleMessages(client)
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Event.New("client disconnected with id \""+client.GetId()+"\" and ip \""+client.GetIp()+"\"", nil).Error())
	}
}

func (server *WebsocketServer) handleMessages(client *WebsocketClient) {
	for {
		messageBytes, err := client.receive()
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Event.New("Failed to receive message from client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error())
			}
			return
		}
		server.incomingMessageCounter.Add(1)
		server.bytesReceivedCounter.Add(uint64(len(messageBytes)))
		if client.rateLimiterBytes != nil && !client.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
			err := client.Send(Message.NewAsync("error", Event.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Event.New("Failed to send rate limit error message to client", err).Error())
				}
			}
			continue
		}
		if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
			err := client.Send(Message.NewAsync("error", Event.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Event.New("Failed to send rate limit error message to client", err).Error())
				}
			}
			continue
		}
		message, err := Message.Deserialize(messageBytes, client.GetId())
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Event.New("Failed to deserialize message from client \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error())
			}
			continue
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
