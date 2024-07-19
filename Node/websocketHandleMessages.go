package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

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
		if time.Since(websocketClient.GetLastMessageTimestamp()) <= time.Duration(node.GetWebsocketComponent().GetWebsocketComponentConfig().ClientMessageCooldownMs)*time.Millisecond {
			err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send rate limit error message to websocket client", err).Error())
				}
			}
			continue
		}
		websocketClient.SetLastMessageTimestamp(time.Now())
		if node.GetWebsocketComponent().GetWebsocketComponentConfig().HandleClientMessagesSequentially {
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
	handler := node.GetWebsocketComponent().GetWebsocketMessageHandlers()[message.GetTopic()]
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
