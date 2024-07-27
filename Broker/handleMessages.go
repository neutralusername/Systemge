package Broker

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func (broker *Broker) handleNodeConnectionMessages(nodeConnection *nodeConnection) {
	if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handling messages from node \""+nodeConnection.name+"\"", nil).Error())
	}
	for broker.isStarted {
		message, err := broker.receive(nodeConnection)
		if err != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from node \""+nodeConnection.name+"\"", err).Error())
			}
			return
		}
		go func() {
			if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
			}
			err = broker.validateMessage(message)
			if err != nil {
				if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Invalid message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", err).Error())
				}
				return
			}
			if message.GetSyncResponseToken() != "" {
				err := broker.handleSyncResponse(message)
				if err != nil {
					if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to handle sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
					}
				} else {
					if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
						infoLogger.Log(Error.New("Received sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
					}
				}
				return
			}
			err = broker.validateTopic(message)
			if err != nil {
				if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Invalid topic for message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", err).Error())
				}
				return
			}
			if message.GetSyncRequestToken() != "" {
				if err := broker.addSyncRequest(nodeConnection, message); err != nil {
					if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to add sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
					}
					//not using handleSyncResponse because the request failed, which means the syncRequest token has not been registered
					err := broker.send(nodeConnection, message.NewResponse("error", broker.node.GetName(), Error.New("sync request failed", err).Error()))
					if err != nil {
						if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
							warningLogger.Log(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
						}
					}
					return
				}
				if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Added sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
				}
			}
			err = broker.handleMessage(nodeConnection, message)
			if err != nil {
				if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", err).Error())
				}
			} else {
				if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
				}
			}
		}()
	}
}

func (broker *Broker) handleMessage(nodeConnection *nodeConnection, message *Message.Message) error {
	switch message.GetTopic() {
	case "unsubscribe":
		err := broker.handleUnsubscribe(nodeConnection, message)
		if err != nil {
			return Error.New("Failed to handle unsubscribe message", err)
		}
	case "subscribe":
		err := broker.handleSubscribe(nodeConnection, message)
		if err != nil {
			return Error.New("Failed to handle subscribe message", err)
		}
	default:
		broker.propagateMessage(message)
	}
	return nil
}

func (broker *Broker) handleSubscribe(nodeConnection *nodeConnection, message *Message.Message) error {
	err := broker.addSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := broker.handleSyncResponse(message.NewResponse("error", broker.node.GetName(), Error.New("failed to add subscription", err).Error()))
		if errResponse != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response for failed subscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", errResponse).Error())
			}
		}
		return Error.New("Failed to add subscription", err)
	}
	err = broker.handleSyncResponse(message.NewResponse("subscribed", broker.node.GetName(), ""))
	if err != nil {
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to send sync response for successful subscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
		}
	}
	return nil
}

func (broker *Broker) handleUnsubscribe(nodeConnection *nodeConnection, message *Message.Message) error {
	err := broker.removeSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := broker.handleSyncResponse(message.NewResponse("error", broker.node.GetName(), Error.New("failed to remove subscription", err).Error()))
		if errResponse != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response for failed unsubscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", errResponse).Error())
			}
		}
		return Error.New("failed to remove subscription", err)
	}
	err = broker.handleSyncResponse(message.NewResponse("unsubscribed", broker.node.GetName(), ""))
	if err != nil {
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to send sync response for successful unsubscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
		}
	}
	return nil
}

func (broker *Broker) propagateMessage(message *Message.Message) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, nodeConnection := range broker.nodeSubscriptions[message.GetTopic()] {
		go func() {
			err := broker.send(nodeConnection, message)
			if err != nil {
				if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to node \""+nodeConnection.name+"\"", err).Error())
				}
				broker.removeNodeConnection(true, nodeConnection)
			}
		}()
	}
}

func (broker *Broker) validateMessage(message *Message.Message) error {
	if message == nil {
		return Error.New("message is nil", nil)
	}
	if message.GetTopic() == "" {
		return Error.New("message topic is empty", nil)
	}
	if message.GetOrigin() == "" {
		return Error.New("message origin is empty", nil)
	}
	if message.GetSyncResponseToken() != "" && message.GetSyncRequestToken() != "" {
		return Error.New("message cannot be both a sync request and a sync response", nil)
	}
	if message.GetSyncResponseToken() != "" {
		return nil
	}
	if broker.config.MaxTopicSize > 0 && len(message.GetTopic()) > broker.config.MaxTopicSize {
		return Error.New("message topic is too long", nil)
	}
	if broker.config.MaxOriginSize > 0 && (len(message.GetOrigin())) > broker.config.MaxOriginSize {
		return Error.New("message origin is too long", nil)
	}
	if broker.config.MaxSyncKeySize > 0 && (len(message.GetSyncRequestToken())) > broker.config.MaxSyncKeySize {
		return Error.New("message sync request token is too long", nil)
	}
	if broker.config.MaxSyncKeySize > 0 && (len(message.GetSyncResponseToken())) > broker.config.MaxSyncKeySize {
		return Error.New("message sync response token is too long", nil)
	}
	if broker.config.MaxPayloadSize > 0 && (len(message.GetPayload())) > broker.config.MaxPayloadSize {
		return Error.New("message payload is too long", nil)
	}

	return nil
}

func (broker *Broker) validateTopic(message *Message.Message) error {
	broker.operationMutex.Lock()
	syncTopicExists := broker.syncTopics[message.GetTopic()]
	asyncTopicExists := broker.asyncTopics[message.GetTopic()]
	broker.operationMutex.Unlock()
	if !syncTopicExists && !asyncTopicExists {
		return Error.New("topic \""+message.GetTopic()+"\" does not exist", nil)
	}
	if syncTopicExists && message.GetSyncRequestToken() == "" {
		return Error.New("topic \""+message.GetTopic()+"\" is a sync topic and message is not a sync request", nil)
	}
	if asyncTopicExists && message.GetSyncRequestToken() != "" {
		return Error.New("topic \""+message.GetTopic()+"\" is an async topic and message is a sync request", nil)
	}
	return nil
}
