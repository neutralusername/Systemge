package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"

	"github.com/gorilla/websocket"
)

func (server *Server) handleWebsocketConnections() {
	for {
		websocketConn := <-server.connChannel
		if websocketConn == nil {
			return
		}
		go server.handleWebsocketConn(websocketConn)
	}
}

func (server *Server) handleWebsocketConn(websocketConn *websocket.Conn) {
	websocketClient := server.addWebsocketConn(websocketConn)
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("websocket client connected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\"", nil).Error())
	}
	if server.onConnectHandler != nil {
		server.onConnectHandler(websocketClient)
	}
	server.handleMessages(websocketClient)
	websocketClient.Disconnect()
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("websocket client disconnected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\"", nil).Error())
	}
}

func (server *Server) handleMessages(websocketClient *Client) {
	for {
		messageBytes, err := websocketClient.receive()
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error())
			}
			return
		}
		server.incomingMessageCounter.Add(1)
		server.bytesReceivedCounter.Add(uint64(len(messageBytes)))
		if websocketClient.rateLimiterBytes != nil && !websocketClient.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		if websocketClient.rateLimiterMsgs != nil && !websocketClient.rateLimiterMsgs.Consume(1) {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		message, err := Message.Deserialize(messageBytes, websocketClient.GetId())
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error())
			}
			continue
		}
		message = Message.NewAsync(message.GetTopic(), message.GetPayload())
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", nil).Error())
		}
		if message.GetTopic() == "heartbeat" {
			server.resetWatchdog(websocketClient)
			continue
		}
		if server.config.HandleClientMessagesSequentially {
			err := server.handleMessage(websocketClient, message)
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
				err := server.handleMessage(websocketClient, message)
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
