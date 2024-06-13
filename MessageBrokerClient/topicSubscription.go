package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
)

// subscribes to a topic from the provided server connection
func (client *Client) subscribeTopic(serverConnection *serverConnection, topic string) error {
	response, err := client.syncMessage(serverConnection, Message.NewSync("subscribe", client.name, topic))
	if err != nil {
		return Error.New("Error sending subscription sync request", err)
	}
	if response.GetTopic() != "subscribed" {
		return Error.New("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	return client.addTopicResolution(topic, serverConnection)
}
