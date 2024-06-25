package Client

import (
	"Systemge/Error"
	"Systemge/Message"
)

// resolves the broker address for the provided topic and sends the async message to the broker responsible for the topic.
func (client *Client) AsyncMessage(topic, origin, payload string) error {
	message := Message.NewAsync(topic, origin, payload)
	if !client.isStarted {
		return Error.New("Client not started", nil)
	}
	brokerConnection, err := client.getBrokerConnectionForTopic(message.GetTopic())
	if err != nil {
		return Error.New("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	return brokerConnection.send(message)
}
