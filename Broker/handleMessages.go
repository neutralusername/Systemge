package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (broker *Broker) handleNodeConnectionMessages(nodeConnection *nodeConnection) {
	for broker.IsStarted() {
		message, err := nodeConnection.receive()
		if err != nil {
			if broker.IsStarted() {
				broker.logger.Log(Error.New("Failed to receive message from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
			}
			broker.disconnect(nodeConnection)
			return
		}
		err = broker.validateMessage(message)
		if err != nil {
			broker.logger.Log(Error.New("Invalid message from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\" \""+string(message.Serialize())+"\"", err).Error())
			broker.disconnect(nodeConnection)
			return
		}
		if message.GetSyncResponseToken() != "" {
			err := broker.handleSyncResponse(message)
			if err != nil {
				broker.logger.Log(Error.New("Failed to handle sync response from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
			}
		} else {
			err = broker.handleMessage(nodeConnection, message)
			if err != nil {
				broker.logger.Log(Error.New("Failed to handle message from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
			}
		}

	}
}

func (broker *Broker) handleMessage(nodeConnection *nodeConnection, message *Message.Message) error {
	if message.GetSyncRequestToken() != "" {
		if err := broker.addSyncRequest(nodeConnection, message); err != nil {
			//not using handleSyncResponse because the request failed, which means the syncRequest token has not been registered
			errSend := nodeConnection.send(message.NewResponse("error", broker.GetName(), Error.New("sync request failed", err).Error()))
			if errSend != nil {
				broker.logger.Log(Error.New("Failed to send error response for failed sync request with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", errSend).Error())
			}
			return Error.New("Failed to add sync request with token \""+message.GetSyncRequestToken()+"\"", err)
		}
	}
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
		errResponse := broker.handleSyncResponse(message.NewResponse("error", broker.GetName(), Error.New("", err).Error()))
		if errResponse != nil {
			broker.logger.Log(Error.New("Failed to send error response for failed subscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", errResponse).Error())
		}
		return Error.New("failed to add subscription", err)
	}
	err = broker.handleSyncResponse(message.NewResponse("subscribed", broker.GetName(), ""))
	if err != nil {
		broker.logger.Log(Error.New("Failed to send sync response for successful subscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
	}
	return nil
}

func (broker *Broker) handleUnsubscribe(nodeConnection *nodeConnection, message *Message.Message) error {
	err := broker.removeSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := broker.handleSyncResponse(message.NewResponse("error", broker.GetName(), Error.New("", err).Error()))
		if errResponse != nil {
			broker.logger.Log(Error.New("Failed to send error response for failed unsubscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", errResponse).Error())
		}
		return Error.New("failed to remove subscription", err)
	}
	err = broker.handleSyncResponse(message.NewResponse("unsubscribed", broker.GetName(), ""))
	if err != nil {
		broker.logger.Log(Error.New("Failed to send sync response for successful unsubscribe request for topic \""+message.GetTopic()+"\" with token \""+message.GetSyncRequestToken()+"\" from node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
	}
	return nil
}

func (broker *Broker) propagateMessage(message *Message.Message) {
	broker.operationMutex.Lock()
	nodes := broker.getSubscribedNodes(message.GetTopic())
	broker.operationMutex.Unlock()
	for _, nodeConnection := range nodes {
		err := nodeConnection.send(message)
		if err != nil {
			broker.logger.Log(Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to node \""+nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
			broker.disconnect(nodeConnection)
		}
	}
}

func (broker *Broker) validateMessage(message *Message.Message) error {
	if message == nil {
		return Error.New("Message is nil", nil)
	}
	if message.GetTopic() == "" {
		return Error.New("Message topic is empty", nil)
	}
	if message.GetOrigin() == "" {
		return Error.New("Message origin is empty", nil)
	}
	if message.GetSyncResponseToken() != "" && message.GetSyncRequestToken() != "" {
		return Error.New("Message cannot be both a sync request and a sync response", nil)
	}
	if message.GetSyncResponseToken() != "" {
		return nil
	}
	broker.operationMutex.Lock()
	syncTopicExists := broker.syncTopics[message.GetTopic()]
	asyncTopicExists := broker.asyncTopics[message.GetTopic()]
	broker.operationMutex.Unlock()
	if !syncTopicExists && !asyncTopicExists {
		return Error.New("Topic \""+message.GetTopic()+"\" does not exist", nil)
	}
	if syncTopicExists && message.GetSyncRequestToken() == "" {
		return Error.New("Topic \""+message.GetTopic()+"\" is a sync topic and message is not a sync request", nil)
	}
	if asyncTopicExists && message.GetSyncRequestToken() != "" {
		return Error.New("Topic \""+message.GetTopic()+"\" is an async topic and message is a sync request", nil)
	}
	return nil
}
