package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

// resolves the broker address for the provided topic and sends the async message to the broker responsible for the topic.
func (node *Node) AsyncMessage(topic, origin, payload string) error {
	message := Message.NewAsync(topic, origin, payload)
	if !node.isStarted {
		return Error.New("Node not started", nil)
	}
	brokerConnection, err := node.getBrokerConnectionForTopic(message.GetTopic())
	if err != nil {
		return Error.New("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	return brokerConnection.send(message)
}
