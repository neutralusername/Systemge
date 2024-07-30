package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func (node *Node) AsyncMessage(topic, payload string) error {
	if systemge := node.systemge; systemge != nil {
		message := Message.NewAsync(topic, payload)
		systemge.outgoingConnectionMutex.Lock()
		for _, outgoingConnection := range systemge.topicResolutions[topic] {
			err := systemge.sendOutgoingConnection(outgoingConnection, message)
			if err != nil {
				if errorLogger := node.GetErrorLogger(); errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send async message with topic \""+topic+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
			}
		}
		systemge.outgoingConnectionMutex.Unlock()
		return nil
	}
	return Error.New("systemge component not initialized", nil)
}
