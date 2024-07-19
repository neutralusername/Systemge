package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tools"
	"time"
)

func (node *Node) subscribeLoop(topic string) {
	for node.IsStarted() {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Attempting subscription to topic \""+topic+"\"", nil).Error())
		}
		if node.subscribeAttempt(topic) {
			break
		}
		time.Sleep(time.Duration(node.GetSystemgeComponent().GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
	}
}

func (node *Node) subscribeAttempt(topic string) bool {
	endpoint, err := node.resolveBrokerForTopic(topic)
	if err != nil {
		if warningLogger := node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to resolve broker for topic \""+topic+"\"", err).Error())
		}
		return false
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Resolved broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
	}
	brokerConnection := node.getBrokerConnection(endpoint.Address)
	if brokerConnection == nil {
		brokerConnection, err = node.connectToBroker(endpoint)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to connect to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", err).Error())
			}
			return false
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Connected to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
		}
		err = node.addBrokerConnection(brokerConnection)
		if err != nil {
			brokerConnection.close()
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to add broker \""+endpoint.Address+"\" for topic \""+topic+"\"", err).Error())
			}
			return false
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Added broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
		}
	} else {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Found existing broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
		}
	}
	err = node.subscribeTopic(brokerConnection, topic)
	if err != nil {
		if warningLogger := node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to subscribe to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", err).Error())
		}
		return false
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Subscribed to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
	}
	err = brokerConnection.addSubscribedTopic(topic)
	if err != nil {
		brokerConnection.mutex.Lock()
		subscribedTopicsCount := len(brokerConnection.subscribedTopics)
		topicResolutionsCount := len(brokerConnection.topicResolutions)
		brokerConnection.mutex.Unlock()
		if subscribedTopicsCount == 0 && topicResolutionsCount == 0 {
			node.handleBrokerDisconnect(brokerConnection)
		}
		if warningLogger := node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to add subscribed topic for broker \""+endpoint.Address+"\" for topic \""+topic+"\"", err).Error())
		}
		return false
	}
	if infoLogger := node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Added subscribed topic for broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
	}
	return true
}

func (node *Node) subscribeTopic(brokerConnection *brokerConnection, topic string) error {
	message := Message.NewSync("subscribe", node.GetName(), topic, node.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC))
	responseChannel, err := node.addMessageWaitingForResponse(message)
	if err != nil {
		return Error.New("failed to add message waiting for response", err)
	}
	err = node.send(brokerConnection, message)
	if err != nil {
		node.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to broker", err)
	}
	response, err := node.receiveSyncResponse(message, responseChannel)
	if err != nil {
		return Error.New("Error sending subscription sync request", err)
	}
	if response.GetTopic() != "subscribed" {
		return Error.New("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	return nil
}
