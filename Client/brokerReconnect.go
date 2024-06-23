package Client

import (
	"Systemge/Utilities"
	"time"
)

func (client *Client) handleBrokerDisconnect(brokerConnection *brokerConnection) {
	removedSubscribedTopics := client.cleanUpDisconnectedBrokerConnection(brokerConnection)
	if len(removedSubscribedTopics) > 0 {
		for _, topic := range removedSubscribedTopics {
			for {
				if !client.IsStarted() {
					return
				}
				client.logger.Log("Attempting reconnect for topic \"" + topic + "\"")
				err := client.attemptToResubscribeToHandlerTopic(topic)
				if err == nil {
					break
				}
				client.logger.Log(Utilities.NewError("Failed reconnect for topic \""+topic+"\"", err).Error())
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func (client *Client) cleanUpDisconnectedBrokerConnection(brokerConnection *brokerConnection) []string {
	client.mapOperationMutex.Lock()
	brokerConnection.mutex.Lock()
	delete(client.activeBrokerConnections, brokerConnection.resolution.GetAddress())
	removedSubscribedTopics := make([]string, 0)
	for topic := range brokerConnection.topics {
		delete(client.topicResolutions, topic)
		if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
			removedSubscribedTopics = append(removedSubscribedTopics, topic)
		}
	}
	brokerConnection.topics = make(map[string]bool)
	brokerConnection.mutex.Unlock()
	client.mapOperationMutex.Unlock()
	return removedSubscribedTopics
}

func (client *Client) attemptToResubscribeToHandlerTopic(topic string) error {
	newBrokerConnection, err := client.getBrokerConnectionForTopic(topic)
	if err != nil {
		return Utilities.NewError("Unable to obtain new broker for topic \""+topic+"\"", err)
	}
	err = client.subscribeTopic(newBrokerConnection, topic)
	if err != nil {
		return Utilities.NewError("Unable to subscribe to topic \""+topic+"\"", err)
	}
	client.logger.Log(Utilities.NewError("Reconnected to broker \""+newBrokerConnection.resolution.GetName()+"\" for topic \""+topic+"\"", nil).Error())
	return nil
}
