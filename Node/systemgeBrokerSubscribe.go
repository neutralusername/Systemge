package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tools"
	"time"
)

func (node *Node) subscribeLoop(topic string) {
	for node.IsStarted() {
		node.GetLogger().Info(Error.New("Attempting subscription to topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
		if node.subscribeAttempt(topic) {
			break
		}
		time.Sleep(time.Duration(node.GetSystemgeComponent().GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
	}
}

func (node *Node) subscribeAttempt(topic string) bool {
	endpoint, err := node.resolveBrokerForTopic(topic)
	if err != nil {
		node.GetLogger().Warning(Error.New("Failed to resolve broker for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
		return false
	}
	node.GetLogger().Info(Error.New("Resolved broker \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
	brokerConnection := node.getBrokerConnection(endpoint.GetAddress())
	if brokerConnection == nil {
		brokerConnection, err = node.connectToBroker(endpoint)
		if err != nil {
			node.GetLogger().Warning(Error.New("Failed to connect to broker \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
			return false
		}
		node.GetLogger().Info(Error.New("Connected to broker \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
		err = node.addBrokerConnection(brokerConnection)
		if err != nil {
			brokerConnection.close()
			node.GetLogger().Warning(Error.New("Failed to add broker connection \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
			return false
		}
		node.GetLogger().Info(Error.New("Added broker connection \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
	} else {
		node.GetLogger().Info(Error.New("Found existing broker connection \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
	}
	err = node.subscribeTopic(brokerConnection, topic)
	if err != nil {
		node.GetLogger().Warning(Error.New("Failed to subscribe to topic \""+topic+"\" on broker \""+endpoint.GetAddress()+"\" for node \""+node.GetName()+"\"", err).Error())
		return false
	}
	node.GetLogger().Info(Error.New("Subscribed to topic \""+topic+"\" on broker \""+endpoint.GetAddress()+"\" for node \""+node.GetName()+"\"", nil).Error())
	err = brokerConnection.addSubscribedTopic(topic)
	if err != nil {
		brokerConnection.mutex.Lock()
		subscribedTopicsCount := len(brokerConnection.subscribedTopics)
		topicResolutionsCount := len(brokerConnection.topicResolutions)
		brokerConnection.mutex.Unlock()
		if subscribedTopicsCount == 0 && topicResolutionsCount == 0 {
			node.handleBrokerDisconnect(brokerConnection)
		}
		node.GetLogger().Warning(Error.New("Failed to add topic \""+topic+"\" to subscribed topics for broker \""+endpoint.GetAddress()+"\" on node \""+node.GetName()+"\"", err).Error())
		return false
	}
	node.GetLogger().Info(Error.New("Added topic \""+topic+"\" to subscribed topics for broker \""+endpoint.GetAddress()+"\" on node \""+node.GetName()+"\"", nil).Error())
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
