package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (node *Node) handleNodeConnectionMessages(nodeConnection *nodeConnection) {
	broker_ := node.broker
	defer broker_.removeNodeConnection(false, nodeConnection)
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handling broker messages from node \""+nodeConnection.name+"\"", nil).Error())
	}
	defer func() {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped handling broker messages from node \""+nodeConnection.name+"\"", nil).Error())
		}
	}()
	for {
		broker := node.broker
		if broker != broker_ {
			return
		}
		message, err := broker.receive(nodeConnection)
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from node \""+nodeConnection.name+"\"", err).Error())
			}
			return
		}
		go func() {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
			}
			err = broker.validateMessage(message)
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Invalid message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", err).Error())
				}
				return
			}
			if message.GetSyncResponseToken() != "" {
				err := broker.handleSyncResponse(message)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to handle sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
					}
				} else {
					if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
						infoLogger.Log(Error.New("Received sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
					}
				}
				return
			}
			err = broker.validateTopic(message)
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Invalid topic for message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", err).Error())
				}
				return
			}
			if message.GetSyncRequestToken() != "" {
				syncRequest, err := broker.addSyncRequest(nodeConnection, message)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to add sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
					}
					//not using handleSyncResponse because adding the request failed, which means the syncRequest token has not been registered
					err := broker.send(nodeConnection, message.NewResponse("error", node.GetName(), Error.New("sync request failed", err).Error()))
					if err != nil {
						if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
							warningLogger.Log(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
						}
					}
					return
				}
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Added sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
				}
				go func() {
					response, err := broker.handleSyncRequest(syncRequest)
					if err != nil {
						if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
							warningLogger.Log(Error.New("Failed to handle sync request with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
						}
						err := broker.send(nodeConnection, message.NewResponse("error", node.GetName(), Error.New("sync request failed", err).Error()))
						if err != nil {
							if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
								warningLogger.Log(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
							}
						}
						return
					} else {
						if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
							infoLogger.Log(Error.New("Handled sync request with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
						}
						err := broker.send(nodeConnection, response)
						if err != nil {
							if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
								warningLogger.Log(Error.New("Failed to send sync response with topic \""+response.GetTopic()+"\" and token \""+response.GetSyncResponseToken()+"\" to node \""+nodeConnection.name+"\"", err).Error())
							}
						} else {
							if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
								infoLogger.Log(Error.New("Sent sync response with topic \""+response.GetTopic()+"\" and token \""+response.GetSyncResponseToken()+"\" to node \""+nodeConnection.name+"\"", nil).Error())
							}
						}
					}
				}()
			}
			err = broker.handleMessage(nodeConnection, message, node.GetName(), node.GetInternalWarningError())
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", err).Error())
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled message with topic \""+message.GetTopic()+"\" from node \""+nodeConnection.name+"\"", nil).Error())
				}
			}
		}()
	}
}

func (broker *brokerComponent) handleMessage(nodeConnection *nodeConnection, message *Message.Message, nodeName string, warningLgger *Tools.Logger) error {
	switch message.GetTopic() {
	case "unsubscribe":
		err := broker.handleUnsubscribe(nodeConnection, message, nodeName, warningLgger)
		if err != nil {
			return Error.New("Failed to handle unsubscribe message", err)
		}
	case "subscribe":
		err := broker.handleSubscribe(nodeConnection, message, nodeName, warningLgger)
		if err != nil {
			return Error.New("Failed to handle subscribe message", err)
		}
	default:
		broker.propagateMessage(message, warningLgger)
	}
	return nil
}

func (broker *brokerComponent) handleSubscribe(nodeConnection *nodeConnection, message *Message.Message, nodeName string, warningLogger *Tools.Logger) error {
	err := broker.addSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := broker.handleSyncResponse(message.NewResponse("error", nodeName, Error.New("failed to add subscription", err).Error()))
		if errResponse != nil {
			if warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response for failed subscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", errResponse).Error())
			}
		}
		return Error.New("Failed to add subscription", err)
	}
	err = broker.handleSyncResponse(message.NewResponse("subscribed", nodeName, ""))
	if err != nil {
		if warningLogger != nil {
			warningLogger.Log(Error.New("Failed to send sync response for successful subscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
		}
	}
	return nil
}

func (broker *brokerComponent) handleUnsubscribe(nodeConnection *nodeConnection, message *Message.Message, nodeName string, warningLogger *Tools.Logger) error {
	err := broker.removeSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := broker.handleSyncResponse(message.NewResponse("error", nodeName, Error.New("failed to remove subscription", err).Error()))
		if errResponse != nil {
			if warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response for failed unsubscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", errResponse).Error())
			}
		}
		return Error.New("failed to remove subscription", err)
	}
	err = broker.handleSyncResponse(message.NewResponse("unsubscribed", nodeName, ""))
	if err != nil {
		if warningLogger != nil {
			warningLogger.Log(Error.New("Failed to send sync response for successful unsubscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\"", err).Error())
		}
	}
	return nil
}

func (broker *brokerComponent) propagateMessage(message *Message.Message, warningLogger *Tools.Logger) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	for _, nodeConnection := range broker.nodeSubscriptions[message.GetTopic()] {
		go func() {
			err := broker.send(nodeConnection, message)
			if err != nil {
				if warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to node \""+nodeConnection.name+"\"", err).Error())
				}
			}
		}()
	}
}

func (broker *brokerComponent) validateMessage(message *Message.Message) error {
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
	if broker.application.GetBrokerComponentConfig().MaxTopicSize > 0 && len(message.GetTopic()) > broker.application.GetBrokerComponentConfig().MaxTopicSize {
		return Error.New("message topic is too long", nil)
	}
	if broker.application.GetBrokerComponentConfig().MaxOriginSize > 0 && (len(message.GetOrigin())) > broker.application.GetBrokerComponentConfig().MaxOriginSize {
		return Error.New("message origin is too long", nil)
	}
	if broker.application.GetBrokerComponentConfig().MaxSyncKeySize > 0 && (len(message.GetSyncRequestToken())) > broker.application.GetBrokerComponentConfig().MaxSyncKeySize {
		return Error.New("message sync request token is too long", nil)
	}
	if broker.application.GetBrokerComponentConfig().MaxSyncKeySize > 0 && (len(message.GetSyncResponseToken())) > broker.application.GetBrokerComponentConfig().MaxSyncKeySize {
		return Error.New("message sync response token is too long", nil)
	}
	if broker.application.GetBrokerComponentConfig().MaxPayloadSize > 0 && (len(message.GetPayload())) > broker.application.GetBrokerComponentConfig().MaxPayloadSize {
		return Error.New("message payload is too long", nil)
	}

	return nil
}

func (broker *brokerComponent) validateTopic(message *Message.Message) error {
	broker.mutex.Lock()
	syncTopicExists := broker.syncTopics[message.GetTopic()]
	asyncTopicExists := broker.asyncTopics[message.GetTopic()]
	broker.mutex.Unlock()
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
