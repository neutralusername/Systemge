package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (node *Node) handleMessages(websocketClient *WebsocketClient) {
	defer websocketClient.Disconnect()
	for node.IsStarted() {
		messageBytes, err := websocketClient.Receive()
		if err != nil {
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			node.logger.Log("Received invalid message from websocketClient \"" + websocketClient.GetId() + "\"" + " with ip \"" + websocketClient.GetIp() + "\"")
			return
		}
		if message.GetTopic() == "heartbeat" {
			websocketClient.ResetWatchdog()
			continue
		}
		if time.Since(websocketClient.GetLastMessageTimestamp()) <= websocketClient.GetMessageCooldown() {
			err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("rate limited", nil).Error()).Serialize())
			if err != nil {
				node.logger.Log(err.Error())
			}
			continue
		}
		websocketClient.SetLastMessageTimestamp(time.Now())
		if websocketClient.GetHandleMessagesConcurrently() {
			go func() {
				err := node.handleWebsocketMessage(websocketClient, message)
				if err != nil {
					node.logger.Log(err.Error())
				}
			}()
		} else {
			err := node.handleWebsocketMessage(websocketClient, message)
			if err != nil {
				node.logger.Log(err.Error())
			}
		}
	}
}

func (node *Node) handleWebsocketMessage(websocketClient *WebsocketClient, message *Message.Message) error {
	message = Message.NewAsync(message.GetTopic(), websocketClient.GetId(), message.GetPayload())
	handler := node.websocketComponent.GetWebsocketMessageHandlers()[message.GetTopic()]
	if handler == nil {
		err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("no handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
		if err != nil {
			return Error.New("failed to send error message", err)
		}
		return nil
	}
	err := handler(node, websocketClient, message)
	if err != nil {
		err := websocketClient.Send(Message.NewAsync("error", node.GetName(), Error.New("error in handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
		if err != nil {
			return Error.New("failed to send error message", err)
		}
	}
	return nil
}
