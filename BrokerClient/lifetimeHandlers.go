package BrokerClient

import (
	"time"

	"github.com/neutralusername/systemge/status"
)

func (messageBrokerClient *Client) handleConnectionLifetime(connection *connection, stopChannel chan bool) {
	select {
	case <-connection.connection.GetCloseChannel():

		messageBrokerClient.mutex.Lock()
		err := connection.connection.StopMessageHandlingLoop()
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(err.Error())
			}
		}

		subscribedAsyncTopicsByClosedConnection := []string{}
		subscribedSyncTopicsByClosedConnection := []string{}
		for topic := range connection.responsibleAsyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleAsyncTopics, topic)
			if messageBrokerClient.subscribedAsyncTopics[topic] {
				subscribedAsyncTopicsByClosedConnection = append(subscribedAsyncTopicsByClosedConnection, topic)
			}
		}
		for topic := range connection.responsibleSyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleSyncTopics, topic)
			if messageBrokerClient.subscribedSyncTopics[topic] {
				subscribedSyncTopicsByClosedConnection = append(subscribedSyncTopicsByClosedConnection, topic)
			}
		}
		delete(messageBrokerClient.brokerConnections, getTcpClientConfigString(connection.tcpClientConfig))
		messageBrokerClient.mutex.Unlock()

		messageBrokerClient.statusMutex.Lock()
		if messageBrokerClient.status != status.Started {
			messageBrokerClient.statusMutex.Unlock()
			return
		}
		for _, topic := range subscribedAsyncTopicsByClosedConnection {
			messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)
		}
		for _, topic := range subscribedSyncTopicsByClosedConnection {
			messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)
		}
		messageBrokerClient.statusMutex.Unlock()
	case <-stopChannel:
		connection.connection.Close()
		messageBrokerClient.mutex.Lock()
		err := connection.connection.StopMessageHandlingLoop()
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(err.Error())
			}
		}
		for topic := range connection.responsibleAsyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleAsyncTopics, topic)
		}
		for topic := range connection.responsibleSyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleSyncTopics, topic)
		}
		delete(messageBrokerClient.brokerConnections, getTcpClientConfigString(connection.tcpClientConfig))
		messageBrokerClient.mutex.Unlock()
	}
}

func (messageBrokerClient *Client) handleTopicResolutionLifetime(topic string, isSynctopic bool, stopChannel chan bool) {
	var topicResolutionTimeout <-chan time.Time
	if messageBrokerClient.config.TopicResolutionLifetimeMs > 0 {
		topicResolutionTimeout = time.After(time.Duration(messageBrokerClient.config.TopicResolutionLifetimeMs) * time.Millisecond)
	}
	select {
	case <-topicResolutionTimeout:
		if (isSynctopic && messageBrokerClient.subscribedSyncTopics[topic]) || (!isSynctopic && messageBrokerClient.subscribedAsyncTopics[topic]) {
			messageBrokerClient.statusMutex.Lock()
			if messageBrokerClient.status != status.Started {
				messageBrokerClient.statusMutex.Unlock()
				return
			}
			messageBrokerClient.startResolutionAttempt(topic, isSynctopic, stopChannel)
			messageBrokerClient.statusMutex.Unlock()
		} else {
			messageBrokerClient.mutex.Lock()
			defer messageBrokerClient.mutex.Unlock()
			for _, connection := range messageBrokerClient.topicResolutions[topic] {
				if isSynctopic {
					delete(connection.responsibleSyncTopics, topic)
				} else {
					delete(connection.responsibleAsyncTopics, topic)
				}
				if len(connection.responsibleAsyncTopics) == 0 && len(connection.responsibleSyncTopics) == 0 {
					delete(messageBrokerClient.brokerConnections, getTcpClientConfigString(connection.tcpClientConfig))
					connection.connection.Close()
				}
			}
			delete(messageBrokerClient.topicResolutions, topic)
		}
	case <-stopChannel:
		return
	}
}
