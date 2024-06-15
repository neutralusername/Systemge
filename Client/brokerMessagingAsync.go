package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
)

// resolves the broker address for the provided topic and sends the async message to the broker responsible for the topic.
func (client *Client) AsyncMessage(topic, origin, payload string) error {
	message := Message.NewAsync(topic, origin, payload)
	if !client.isStarted {
		return Utilities.NewError("Client not connected", nil)
	}
	brokerConnection, err := client.getBrokerConnectionForTopic(message.GetTopic())
	if err != nil {
		return Utilities.NewError("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	return brokerConnection.send(message)
}
