package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// AsyncMessage sends an async message.
// If receiverNames is empty, the message will be sent to all connected servers that are interested in the topic.
// If receiverNames are specified, the message will be sent to the servers with the specified names.
// Blocking until all messages are sent
func (client *SystemgeClient) AsyncMessage(topic, payload string, receiverNames ...string) error {
	client.createMessageTaskGroup(Message.NewAsync(topic, payload), receiverNames...).ExecuteTasks()
	return nil
}

// Alternative call for AsyncMessage with a config struct instead of multiple arguments
func (client *SystemgeClient) AsyncMessage_(config *Config.Message) error {
	return client.AsyncMessage(config.Topic, config.Payload, config.Receivernames...)
}

// SyncMessage sends a sync message.
// If receiverNames is empty, the message will be sent to all connected servers that are interested in the topic.
// If receiverNames are specified, the message will be sent to the servers with the specified names.
// Blocking until all requests are sent
func (client *SystemgeClient) SyncMessage(topic, payload string, receiverNames ...string) (*SyncResponseChannel, error) {
	responseChannel := client.addResponseChannel(client.randomizer, topic, payload)
	waitgroup := client.createMessageTaskGroup(responseChannel.GetRequestMessage(), receiverNames...)
	responseChannel.responseChannel = make(chan *Message.Message, waitgroup.GetTaskCount())
	if cap(responseChannel.responseChannel) == 0 {
		responseChannel.Close()
	}
	waitgroup.ExecuteTasks()
	go client.responseChannelTimeout(client.stopChannel, responseChannel)
	return responseChannel, nil
}

// Alternative call for SyncMessage with a config struct instead of multiple arguments
func (client *SystemgeClient) SyncMessage_(config *Config.Message) (*SyncResponseChannel, error) {
	return client.SyncMessage(config.Topic, config.Payload, config.Receivernames...)
}

func (client *SystemgeClient) createMessageTaskGroup(message *Message.Message, receiverNames ...string) *Tools.TaskGroup {
	waitgroup := Tools.NewTaskGroup()
	client.serverConnectionMutex.RLock()
	defer client.serverConnectionMutex.RUnlock()
	if len(receiverNames) == 0 {
		for _, serverConnection := range client.topicResolutions[message.GetTopic()] {
			waitgroup.AddTask(func() {
				err := client.messageServerConnection(serverConnection, message)
				if err != nil {
					if errorLogger := client.errorLogger; errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to server connection \""+serverConnection.name+"\"", err).Error())
					}
					if mailer := client.mailer; mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to server connection \""+serverConnection.name+"\"", err).Error()))
						if err != nil {
							if errorLogger := client.errorLogger; errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				} else {
					if infoLogger := client.infoLogger; infoLogger != nil {
						infoLogger.Log("Sent message with topic \"" + message.GetTopic() + "\" to server connection \"" + serverConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
					}
				}
			})
		}
	} else {
		for _, receiverName := range receiverNames {
			if serverConnection := client.topicResolutions[message.GetTopic()][receiverName]; serverConnection != nil {
				waitgroup.AddTask(func() {
					err := client.messageServerConnection(serverConnection, message)
					if err != nil {
						if errorLogger := client.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to server connection \""+serverConnection.name+"\"", err).Error())
						}
						if mailer := client.mailer; mailer != nil {
							err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to server connection \""+serverConnection.name+"\"", err).Error()))
							if err != nil {
								if errorLogger := client.errorLogger; errorLogger != nil {
									errorLogger.Log(Error.New("Failed sending mail", err).Error())
								}
							}
						}
					} else {
						if infoLogger := client.infoLogger; infoLogger != nil {
							infoLogger.Log("Sent message with topic \"" + message.GetTopic() + "\" to server connection \"" + serverConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
						}
					}
				})
			} else {
				if errorLogger := client.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\". No server connection with name \""+receiverName+"\" found", nil).Error())
				}
				if mailer := client.mailer; mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send message with topic \""+message.GetTopic()+"\". No server connection with name \""+receiverName+"\" found", nil).Error()))
					if err != nil {
						if errorLogger := client.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			}
		}
	}
	return waitgroup
}
