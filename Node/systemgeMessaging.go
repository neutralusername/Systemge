package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// AsyncMessage sends an async message.
// If receiverNames is empty, the message will be sent to all connected nodes that are interested in the topic.
// If receiverNames are specified, the message will be sent to the nodes with the specified names.
// Blocking until all messages are sent
func (node *Node) AsyncMessage(topic, payload string, receiverNames ...string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.createOutgoingMessageWaitgroup(node.GetErrorLogger(), node.GetInternalInfoLogger(), Message.NewAsync(topic, payload), receiverNames...).Execute()
		return nil
	}
	return Error.New("systemge component not initialized", nil)
}

// SyncMessage sends a sync message.
// If receiverNames is empty, the message will be sent to all connected nodes that are interested in the topic.
// If receiverNames are specified, the message will be sent to the nodes with the specified names.
// Blocking until all requests are sent
func (node *Node) SyncMessage(topic, payload string, receiverNames ...string) (*SyncResponseChannel, error) {
	if systemge := node.systemge; systemge != nil {
		responseChannel := systemge.addResponseChannel(node.randomizer, topic, payload, systemge.config.SyncResponseLimit)
		systemge.createOutgoingMessageWaitgroup(node.GetErrorLogger(), node.GetInternalInfoLogger(), responseChannel.GetRequestMessage(), receiverNames...).Execute()
		go systemge.responseChannelTimeout(node.stopChannel, responseChannel)
		return responseChannel, nil
	}
	return nil, Error.New("systemge not initialized", nil)
}

func (systemge *systemgeComponent) createOutgoingMessageWaitgroup(errorLogger *Tools.Logger, infoLogger *Tools.Logger, message *Message.Message, receiverNames ...string) *Tools.Waitgroup {
	waitgroup := Tools.NewWaitgroup()
	systemge.outgoingConnectionMutex.Lock()
	defer systemge.outgoingConnectionMutex.Unlock()
	if len(receiverNames) == 0 {
		for _, outgoingConnection := range systemge.topicResolutions[message.GetTopic()] {
			waitgroup.Add(func() {
				err := systemge.messageOutgoingConnection(outgoingConnection, message)
				if err != nil {
					if errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
					}
				} else {
					if infoLogger != nil {
						infoLogger.Log("Sent message with topic \"" + message.GetTopic() + "\" to outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
					}
				}
			})
		}
	} else {
		for _, receiverName := range receiverNames {
			if outgoingConnection := systemge.topicResolutions[message.GetTopic()][receiverName]; outgoingConnection != nil {
				waitgroup.Add(func() {
					err := systemge.messageOutgoingConnection(outgoingConnection, message)
					if err != nil {
						if errorLogger != nil {
							errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
						}
					} else {
						if infoLogger != nil {
							infoLogger.Log("Sent message with topic \"" + message.GetTopic() + "\" to outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
						}
					}
				})
			} else {
				if errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\". No outgoing node connection with name \""+receiverName+"\" found", nil).Error())
				}
			}
		}
	}
	return waitgroup
}
