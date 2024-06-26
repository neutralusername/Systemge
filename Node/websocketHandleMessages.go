package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (client *Node) handleMessages(websocketClient *WebsocketClient) {
	defer websocketClient.Disconnect()
	for client.IsStarted() {
		messageBytes, err := websocketClient.Receive()
		if err != nil {
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			client.logger.Log("Received invalid message from client \"" + websocketClient.GetId() + "\"" + " with ip \"" + websocketClient.GetIp() + "\"")
			return
		}
		if message.GetTopic() == "heartbeat" {
			websocketClient.ResetWatchdog()
			continue
		}
		if time.Since(websocketClient.GetLastMessageTimestamp()) <= websocketClient.GetMessageCooldown() {
			continue
		}
		websocketClient.SetLastMessageTimestamp(time.Now())
		if websocketClient.GetHandleMessagesConcurrently() {
			go func() {
				err := client.handleWebsocketMessage(websocketClient, message)
				if err != nil {
					client.logger.Log(err.Error())
				}
			}()
		} else {
			err := client.handleWebsocketMessage(websocketClient, message)
			if err != nil {
				client.logger.Log(err.Error())
			}
		}
	}
}

func (client *Node) handleWebsocketMessage(websocketClient *WebsocketClient, message *Message.Message) error {
	message = Message.NewAsync(message.GetTopic(), websocketClient.GetId(), message.GetPayload())
	handler := client.websocketApplication.GetWebsocketMessageHandlers()[message.GetTopic()]
	if handler == nil {
		err := websocketClient.Send(Message.NewAsync("error", websocketClient.GetId(), Error.New("no handler for topic \""+message.GetTopic()+"\" from client \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
		if err != nil {
			return Error.New("failed to send error message", err)
		}
		return nil
	}
	err := handler(client, websocketClient, message)
	if err != nil {
		err := websocketClient.Send(Message.NewAsync("error", websocketClient.GetId(), Error.New("error in handler for topic \""+message.GetTopic()+"\" from client \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
		if err != nil {
			return Error.New("failed to send error message", err)
		}
	}
	return nil
}
