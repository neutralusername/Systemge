package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"

	"github.com/gorilla/websocket"
)

func (websocket *websocketComponent) handleWebsocketConnections() {
	websocketConn := <-websocket.connChannel
	if websocketConn == nil {
		return
	}
	go websocket.handleWebsocketConn(websocketConn)
}

func (websocket *websocketComponent) handleWebsocketConn(websocketConn *websocket.Conn) {
	websocketClient := websocket.addWebsocketConn(websocketConn)
	if infoLogger := websocket.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("websocket client connected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\"", nil).Error())
	}
	websocket.onConnectHandler(websocketClient)
	websocket.handleMessages(websocketClient)
	websocketClient.Disconnect()
	if infoLogger := websocket.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("websocket client disconnected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\"", nil).Error())
	}
}

func (websocket *websocketComponent) handleMessages(websocketClient *WebsocketClient) {
	for {
		messageBytes, err := websocketClient.receive()
		if err != nil {
			if warningLogger := websocket.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error())
			}
			return
		}
		websocket.incomingMessageCounter.Add(1)
		websocket.bytesReceivedCounter.Add(uint64(len(messageBytes)))
		if websocketClient.rateLimiterBytes != nil && !websocketClient.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := websocket.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		if websocketClient.rateLimiterMsgs != nil && !websocketClient.rateLimiterMsgs.Consume(1) {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := websocket.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		message, err := Message.Deserialize(messageBytes, websocketClient.GetId())
		if err != nil {
			if warningLogger := websocket.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error())
			}
			continue
		}
		message = Message.NewAsync(message.GetTopic(), message.GetPayload())
		if infoLogger := websocket.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", nil).Error())
		}
		if message.GetTopic() == "heartbeat" {
			websocket.resetWatchdog(websocketClient)
			continue
		}
		if websocket.config.HandleClientMessagesSequentially {
			err := websocket.handleMessage(websocketClient, message)
			if err != nil {
				if warningLogger := websocket.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle message (sequentially)", err).Error())
				}
			} else {
				if infoLogger := websocket.infoLogger; infoLogger != nil {
					infoLogger.Log(Error.New("Handled message (sequentially)", nil).Error())
				}
			}
		} else {
			go func() {
				err := websocket.handleMessage(websocketClient, message)
				if err != nil {
					if warningLogger := websocket.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("Failed to handle message (concurrently)", err).Error())
					}
				} else {
					if infoLogger := websocket.infoLogger; infoLogger != nil {
						infoLogger.Log(Error.New("Handled message (concurrently)", nil).Error())
					}
				}
			}()
		}
	}
}
