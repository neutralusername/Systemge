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
		err = brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send async message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
			}
			return Error.New("failed sending async message", err)
		}
		systemge.outgoingAsyncMessageCounter.Add(1)
		return nil
	}
	return Error.New("systemge component not initialized", nil)
}
