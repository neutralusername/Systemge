package Node

import (
	"Systemge/Error"
	"time"
)

func (node *Node) handleBrokerDisconnect(brokerConnection *brokerConnection) {
	removedSubscribedTopics := node.cleanUpDisconnectedBrokerConnection(brokerConnection)
	for _, topic := range removedSubscribedTopics {
		for {
			if !node.IsStarted() {
				return
			}
			node.logger.Log("Attempting reconnect for topic \"" + topic + "\"")
			err := node.attemptToResubscribeToHandlerTopic(topic)
			if err == nil {
				break
			}
			node.logger.Log(Error.New("Failed reconnect for topic \""+topic+"\"", err).Error())
			time.Sleep(1 * time.Second)
		}
	}
}

func (node *Node) cleanUpDisconnectedBrokerConnection(brokerConnection *brokerConnection) []string {
	node.mutex.Lock()
	brokerConnection.mutex.Lock()
	delete(node.activeBrokerConnections, brokerConnection.resolution.GetAddress())
	removedSubscribedTopics := make([]string, 0)
	for topic := range brokerConnection.topics {
		delete(node.topicResolutions, topic)
		if node.application.GetAsyncMessageHandlers()[topic] != nil || node.application.GetSyncMessageHandlers()[topic] != nil {
			removedSubscribedTopics = append(removedSubscribedTopics, topic)
		}
	}
	brokerConnection.topics = make(map[string]bool)
	brokerConnection.mutex.Unlock()
	node.mutex.Unlock()
	return removedSubscribedTopics
}

func (node *Node) attemptToResubscribeToHandlerTopic(topic string) error {
	newBrokerConnection, err := node.getBrokerConnectionForTopic(topic)
	if err != nil {
		return Error.New("Unable to obtain new broker for topic \""+topic+"\"", err)
	}
	err = node.subscribeTopic(newBrokerConnection, topic)
	if err != nil {
		return Error.New("Unable to subscribe to topic \""+topic+"\"", err)
	}
	node.logger.Log(Error.New("Reconnected to broker \""+newBrokerConnection.resolution.GetName()+"\" for topic \""+topic+"\"", nil).Error())
	return nil
}
