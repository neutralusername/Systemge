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
			go func() {
				err := systemge.messageOutgoingConnection(outgoingConnection, message)
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send async message with topic \""+topic+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
					}
				} else {
					if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
						infoLogger.Log("Sent async message with topic \"" + topic + "\" to outgoing node connection \"" + outgoingConnection.name + "\"")
					}
				}
			}()
		}
		systemge.outgoingConnectionMutex.Unlock()
		return nil
	}
	return Error.New("systemge component not initialized", nil)
}
