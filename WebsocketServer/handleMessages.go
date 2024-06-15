package WebsocketServer

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"Systemge/WebsocketClient"
	"time"
)

func (server *Server) handleMessages(client *WebsocketClient.Client) {
	defer client.Disconnect()
	for server.IsStarted() {
		messageBytes, err := client.Receive()
		if err != nil {
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			server.logger.Log("Received invalid message from client \"" + client.GetId() + "\"" + " with ip \"" + client.GetIp() + "\"")
			return
		}
		if message.GetTopic() == "heartbeat" {
			client.ResetWatchdog()
			continue
		}
		if time.Since(client.GetLastMessageTimestamp()) <= client.GetMessageCooldown() {
			continue
		}
		client.SetLastMessageTimestamp(time.Now())
		if client.GetHandleMessagesConcurrently() {
			go func() {
				err := server.handleMessage(client, message)
				if err != nil {
					server.logger.Log(err.Error())
				}
			}()
		} else {
			err := server.handleMessage(client, message)
			if err != nil {
				server.logger.Log(err.Error())
			}
		}
	}
}

func (server *Server) handleMessage(websocketClient *WebsocketClient.Client, message *Message.Message) error {
	message = Message.NewAsync(message.GetTopic(), websocketClient.GetId(), message.GetPayload())
	messageHandler := server.websocketApplication.GetWebsocketMessageHandlers()[message.GetTopic()]
	if messageHandler == nil {
		return Utilities.NewError("no handler for topic \""+message.GetTopic()+"\" from client \""+websocketClient.GetId()+"\"", nil)
	}
	err := server.websocketApplication.GetWebsocketMessageHandlers()[message.GetTopic()](websocketClient, message)
	if err != nil {
		websocketClient.Send(Message.NewAsync("error", websocketClient.GetId(), Utilities.NewError("error in handler for topic \""+message.GetTopic()+"\" from client \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
	}
	return nil
}
