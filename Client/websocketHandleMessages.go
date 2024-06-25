package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"Systemge/WebsocketClient"
	"time"
)

func (client *Client) handleMessages(websocketClient *WebsocketClient.Client) {
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

func (client *Client) handleWebsocketMessage(websocketClient *WebsocketClient.Client, message *Message.Message) error {
	message = Message.NewAsync(message.GetTopic(), websocketClient.GetId(), message.GetPayload())
	handler := client.websocketApplication.GetWebsocketMessageHandlers()[message.GetTopic()]
	if handler == nil {
		err := websocketClient.Send(Message.NewAsync("error", websocketClient.GetId(), Utilities.NewError("no handler for topic \""+message.GetTopic()+"\" from client \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
		if err != nil {
			return Utilities.NewError("failed to send error message", err)
		}
		return nil
	}
	err := handler(client, websocketClient, message)
	if err != nil {
		err := websocketClient.Send(Message.NewAsync("error", websocketClient.GetId(), Utilities.NewError("error in handler for topic \""+message.GetTopic()+"\" from client \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
		if err != nil {
			return Utilities.NewError("failed to send error message", err)
		}
	}
	return nil
}
