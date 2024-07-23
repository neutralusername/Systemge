package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

// resolves the broker address for the provided topic and sends the async message to the broker responsible for the topic.
func (node *Node) AsyncMessage(topic, origin, payload string) error {
	if systemge := node.systemge; systemge != nil {
		message := Message.NewAsync(topic, origin, payload)
		if !node.IsStarted() {
			return Error.New("node not started", nil)
		}
		brokerConnection, err := node.getBrokerConnectionForTopic(message.GetTopic())
		if err != nil {
			return Error.New("failed getting broker connection", err)
		}
		systemge.outgoingAsyncMessageCounter.Add(1)
		return brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message)
	}
	return Error.New("systemge component not initialized", nil)
}
