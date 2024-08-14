package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"

	"github.com/gorilla/websocket"
)

func (server *WebsocketServer) handleWebsocketConnections() {
	for {
		websocketConn := <-server.connChannel
		if websocketConn == nil {
			return
		}
		go server.handleWebsocketConn(websocketConn)
	}
}

func (server *WebsocketServer) handleWebsocketConn(websocketConn *websocket.Conn) {
	client := server.addWebsocketConn(websocketConn)
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("websocket client connected with id \""+client.GetId()+"\" and ip \""+client.GetIp()+"\"", nil).Error())
	}
	if server.onConnectHandler != nil {
		server.onConnectHandler(client)
	}
	server.handleMessages(client)
	client.Disconnect()
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("websocket client disconnected with id \""+client.GetId()+"\" and ip \""+client.GetIp()+"\"", nil).Error())
	}
}

func (server *WebsocketServer) handleMessages(client *Client) {
	for {
		messageBytes, err := client.receive()
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from websocketClient \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error())
			}
			return
		}
		server.incomingMessageCounter.Add(1)
		server.bytesReceivedCounter.Add(uint64(len(messageBytes)))
		if client.rateLimiterBytes != nil && !client.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
			err := client.Send(Message.NewAsync("error", Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		if client.rateLimiterMsgs != nil && !client.rateLimiterMsgs.Consume(1) {
			err := client.Send(Message.NewAsync("error", Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		message, err := Message.Deserialize(messageBytes, client.GetId())
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message from websocketClient \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", err).Error())
			}
			continue
		}
		message = Message.NewAsync(message.GetTopic(), message.GetPayload())
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from websocketClient \""+client.GetId()+"\" with ip \""+client.GetIp()+"\"", nil).Error())
		}
		if message.GetTopic() == "heartbeat" {
			server.resetWatchdog(client)
			continue
		}
		if server.config.HandleClientMessagesSequentially {
			err := server.handleWebsocketMessage(client, message)
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle message (sequentially)", err).Error())
				}
			} else {
				if infoLogger := server.infoLogger; infoLogger != nil {
					infoLogger.Log(Error.New("Handled message (sequentially)", nil).Error())
				}
			}
		} else {
			go func() {
				err := server.handleWebsocketMessage(client, message)
				if err != nil {
					if warningLogger := server.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("Failed to handle message (concurrently)", err).Error())
					}
				} else {
					if infoLogger := server.infoLogger; infoLogger != nil {
						infoLogger.Log(Error.New("Handled message (concurrently)", nil).Error())
					}
				}
			}()
		}
	}
}

func (server *WebsocketServer) handleWebsocketMessage(websocketClient *Client, message *Message.Message) error {
	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()
	if handler == nil {
		err := websocketClient.Send(Message.NewAsync("error", Error.New("no handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
			}
		}
		return Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
	}
	err := handler(websocketClient, message)
	if err != nil {
		err := websocketClient.Send(Message.NewAsync("error", Error.New("error in handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
			}
		}
	}
	return nil
}
