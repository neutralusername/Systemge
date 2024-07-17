package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (node *Node) handleMessages(websocketClient *WebsocketClient) {
	defer websocketClient.Disconnect()
	for node.IsStarted() {
		message, err := websocketClient.Receive()
		if err != nil {
			node.config.Logger.Warning(Error.New("failed to receive message from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\" on node \""+node.GetName()+"\"", err).Error())
			return
		} else {
			node.config.Logger.Info(Error.New("received message with topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\" with ip \""+websocketClient.GetIp()+"\" on node \""+node.GetName()+"\"", nil).Error())
		}
		if message.GetTopic() == "heartbeat" {
			node.ResetWatchdog(websocketClient)
			continue
		}
		if time.Since(websocketClient.GetLastMessageTimestamp()) <= time.Duration(node.GetWebsocketComponent().GetWebsocketComponentConfig().ClientMessageCooldownMs)*time.Millisecond {
			err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				node.config.Logger.Warning(Error.New("failed to send rate limit error message to websocket client on node \""+node.GetName()+"\"", err).Error())
			}
			continue
		}
		websocketClient.SetLastMessageTimestamp(time.Now())
		if node.GetWebsocketComponent().GetWebsocketComponentConfig().HandleClientMessagesSequentially {
			err := node.handleWebsocketMessage(websocketClient, message)
			if err != nil {
				node.config.Logger.Warning(Error.New("failed to handle message (sequentially) on node \""+node.GetName()+"\"", err).Error())
			} else {
				node.config.Logger.Info(Error.New("handled message (sequentially) on node \""+node.GetName()+"\"", nil).Error())
			}
		} else {
			go func() {
				err := node.handleWebsocketMessage(websocketClient, message)
				if err != nil {
					node.config.Logger.Warning(Error.New("failed to handle message (concurrently) on node \""+node.GetName()+"\"", err).Error())
				} else {
					node.config.Logger.Info(Error.New("handled message (concurrently) on node \""+node.GetName()+"\"", nil).Error())
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
			node.config.Logger.Warning(Error.New("failed to send error message to websocket client", err).Error())
		}
		return Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
	}
	err := handler(node, websocketClient, message)
	if err != nil {
		err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("error in handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
		if err != nil {
			node.config.Logger.Warning(Error.New("failed to send error message to websocket client", err).Error())
		}
	}
	return nil
}
