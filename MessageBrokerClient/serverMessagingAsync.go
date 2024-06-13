package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
)

// resolves the broker address for the provided topic and sends the async message to the broker responsible for the topic.
func (client *Client) AsyncMessage(message *Message.Message) error {
	if !client.isStarted {
		return Error.New("Client not connected", nil)
	}
	serverConnection, err := client.getServerConnectionForTopic(message.GetTopic())
	if err != nil {
		return Error.New("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	return serverConnection.send(message)
}
