package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (node *Node) AsyncMessage(topic, payload string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.createOutgoingMessageWaitgroup(node.GetErrorLogger(), node.GetInternalInfoLogger(), Message.NewAsync(topic, payload)).Execute()
		return nil
	}
	return Error.New("systemge component not initialized", nil)
}

func (node *Node) SyncMessage(topic, payload string) (*SyncResponseChannel, error) {
	if systemge := node.systemge; systemge != nil {
		responseChannel := systemge.addResponseChannel(node.randomizer, topic, payload, systemge.config.SyncResponseLimit)
		systemge.createOutgoingMessageWaitgroup(node.GetErrorLogger(), node.GetInternalInfoLogger(), responseChannel.GetRequestMessage()).Execute()
		go systemge.responseChannelTimeout(node.stopChannel, responseChannel)
		return responseChannel, nil
	}
	return nil, Error.New("systemge not initialized", nil)
}

func (systemge *systemgeComponent) createOutgoingMessageWaitgroup(errorLogger *Tools.Logger, infoLogger *Tools.Logger, message *Message.Message) *Tools.Waitgroup {
	waitgroup := Tools.NewWaitgroup()
	systemge.outgoingConnectionMutex.Lock()
	defer systemge.outgoingConnectionMutex.Unlock()
	for _, outgoingConnection := range systemge.topicResolutions[message.GetTopic()] {
		waitgroup.Add(func() {
			err := systemge.messageOutgoingConnection(outgoingConnection, message)
			if err != nil {
				if errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send sync request with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
			} else {
				if infoLogger != nil {
					infoLogger.Log("Sent sync request with topic \"" + message.GetTopic() + "\" to outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
				}
			}
		})
	}
	return waitgroup
}
