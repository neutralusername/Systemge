package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

// resolves the broker address for the provided topic and sends the async message to the broker responsible for the topic.
func (node *Node) AsyncMessage(topic, origin, payload string) error {
	message := Message.NewAsync(topic, origin, payload)
	if !node.IsStarted() {
		return Error.New("node not started", nil)
	}
	brokerConnection := node.getBrokerConnectionForTopic(message.GetTopic())
	if brokerConnection == nil {
		return Error.New("failed resolving broker address for topic \""+message.GetTopic()+"\"", nil)
	}
	return node.send(brokerConnection, message)
}
