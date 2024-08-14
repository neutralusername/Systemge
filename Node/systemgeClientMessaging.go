package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// AsyncMessage sends an async message.
// If receiverNames is empty, the message will be sent to all connected nodes that are interested in the topic.
// If receiverNames are specified, the message will be sent to the nodes with the specified names.
// Blocking until all messages are sent
func (node *Node) AsyncMessage(topic, payload string, receiverNames ...string) error {
	if systemge := node.systemgeClient; systemge != nil {
		node.createOutgoingMessageTaskGroup(systemge, Message.NewAsync(topic, payload), receiverNames...).ExecuteTasks()
		return nil
	}
	return Error.New("systemge component not initialized", nil)
}

// Alternative call for AsyncMessage with a config struct instead of multiple arguments
func (node *Node) AsyncMessage_(config *Config.Message) error {
	return node.AsyncMessage(config.Topic, config.Payload, config.NodeNames...)
}

// SyncMessage sends a sync message.
// If receiverNames is empty, the message will be sent to all connected nodes that are interested in the topic.
// If receiverNames are specified, the message will be sent to the nodes with the specified names.
// Blocking until all requests are sent
func (node *Node) SyncMessage(topic, payload string, receiverNames ...string) (*SyncResponseChannel, error) {
	if systemge := node.systemgeClient; systemge != nil {
		responseChannel := systemge.addResponseChannel(node.randomizer, topic, payload)
		waitgroup := node.createOutgoingMessageTaskGroup(systemge, responseChannel.GetRequestMessage(), receiverNames...)
		responseChannel.responseChannel = make(chan *Message.Message, waitgroup.GetTaskCount())
		if cap(responseChannel.responseChannel) == 0 {
			responseChannel.Close()
		}
		waitgroup.ExecuteTasks()
		go systemge.responseChannelTimeout(systemge.stopChannel, responseChannel)
		return responseChannel, nil
	}
	return nil, Error.New("systemge not initialized", nil)
}

// Alternative call for SyncMessage with a config struct instead of multiple arguments
func (node *Node) SyncMessage_(config *Config.Message) (*SyncResponseChannel, error) {
	return node.SyncMessage(config.Topic, config.Payload, config.NodeNames...)
}

func (node *Node) createOutgoingMessageTaskGroup(systemge *systemgeClientComponent, message *Message.Message, receiverNames ...string) *Tools.TaskGroup {
	waitgroup := Tools.NewTaskGroup()
	systemge.outgoingConnectionMutex.RLock()
	defer systemge.outgoingConnectionMutex.RUnlock()
	if len(receiverNames) == 0 {
		for _, outgoingConnection := range systemge.topicResolutions[message.GetTopic()] {
			waitgroup.AddTask(func() {
				err := systemge.messageOutgoingConnection(outgoingConnection, message)
				if err != nil {
					if errorLogger := node.errorLogger; errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
					}
					if mailer := node.mailer; mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error()))
						if err != nil {
							if errorLogger := node.errorLogger; errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				} else {
					if infoLogger := node.infoLogger; infoLogger != nil {
						infoLogger.Log("Sent message with topic \"" + message.GetTopic() + "\" to outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
					}
				}
			})
		}
	} else {
		for _, receiverName := range receiverNames {
			if outgoingConnection := systemge.topicResolutions[message.GetTopic()][receiverName]; outgoingConnection != nil {
				waitgroup.AddTask(func() {
					err := systemge.messageOutgoingConnection(outgoingConnection, message)
					if err != nil {
						if errorLogger := node.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
						}
						if mailer := node.mailer; mailer != nil {
							err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error()))
							if err != nil {
								if errorLogger := node.errorLogger; errorLogger != nil {
									errorLogger.Log(Error.New("Failed sending mail", err).Error())
								}
							}
						}
					} else {
						if infoLogger := node.infoLogger; infoLogger != nil {
							infoLogger.Log("Sent message with topic \"" + message.GetTopic() + "\" to outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
						}
					}
				})
			} else {
				if errorLogger := node.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\". No outgoing node connection with name \""+receiverName+"\" found", nil).Error())
				}
				if mailer := node.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send message with topic \""+message.GetTopic()+"\". No outgoing node connection with name \""+receiverName+"\" found", nil).Error()))
					if err != nil {
						if errorLogger := node.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}
	}
	return waitgroup
}
