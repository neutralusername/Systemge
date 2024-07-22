package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"

	"github.com/gorilla/websocket"
)

func (node *Node) handleWebsocketConnections() {
	for websocket := node.websocket; websocket != nil; {
		websocketConn := <-websocket.connChannel
		if websocketConn == nil {
			return
		}
		go node.handleWebsocketConn(websocketConn)
	}
}

func (node *Node) handleWebsocketConn(websocketConn *websocket.Conn) {
	websocketClient := node.websocket.addWebsocketConn(websocketConn)
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("websocket client connected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\"", nil).Error())
	}
	node.websocket.application.OnConnectHandler(node, websocketClient)
	node.handleMessages(websocketClient)
	websocketClient.Disconnect()
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("websocket client disconnected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\"", nil).Error())
	}
}

func (node *Node) handleMessages(websocketClient *WebsocketClient) {
	for node.IsStarted() {
		message, err := websocketClient.Receive()
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", err).Error())
			}
			return
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\"", nil).Error())
		}
		if message.GetTopic() == "heartbeat" {
			node.ResetWatchdog(websocketClient)
			continue
		}
		if time.Since(websocketClient.GetLastMessageTimestamp()) <= time.Duration(node.websocket.application.GetWebsocketComponentConfig().ClientMessageCooldownMs)*time.Millisecond {
			err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		websocketClient.SetLastMessageTimestamp(time.Now())
		if node.websocket.application.GetWebsocketComponentConfig().HandleClientMessagesSequentially {
			err := node.handleWebsocketMessage(websocketClient, message)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle message (sequentially)", err).Error())
				}
			} else {
				if infoLogger := node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled message (sequentially)", nil).Error())
				}
			}
		} else {
			go func() {
				err := node.handleWebsocketMessage(websocketClient, message)
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to handle message (concurrently)", err).Error())
					}
				} else {
					if infoLogger := node.GetInfoLogger(); infoLogger != nil {
						infoLogger.Log(Error.New("Handled message (concurrently)", nil).Error())
					}
				}
			}()
		}
	}
}

func (node *Node) handleWebsocketMessage(websocketClient *WebsocketClient, message *Message.Message) error {
	handler := node.websocket.application.GetWebsocketMessageHandlers()[message.GetTopic()]
	if handler == nil {
		err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("no handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
			}
		}
		return Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
	}
	err := handler(node, websocketClient, message)
	if err != nil {
		err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("error in handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
			}
		}
	}
	return nil
}
