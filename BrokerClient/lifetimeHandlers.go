package BrokerClient

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (messageBrokerClient *Client) handleConnectionLifetime(connection *connection, stopChannel chan bool) {
	select {
	case <-connection.connection.GetCloseChannel():
		messageBrokerClient.mutex.Lock()
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
		delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
		messageBrokerClient.mutex.Unlock()

		messageBrokerClient.statusMutex.Lock()
		if messageBrokerClient.status != Status.STARTED {
			messageBrokerClient.statusMutex.Unlock()
			return
		}
		for _, topic := range subscribedAsyncTopicsByClosedConnection {
			err := messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)
			if err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to restart resolution attempt for topic \""+topic+"\"", err).Error())
				}
				if messageBrokerClient.mailer != nil {
					if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to restart resolution attempt", err).Error())); err != nil {
						if messageBrokerClient.errorLogger != nil {
							messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
						}
					}
				}
			}
		}
		for _, topic := range subscribedSyncTopicsByClosedConnection {
			err := messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)
			if err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to restart resolution attempt for topic \""+topic+"\"", err).Error())
				}
				if messageBrokerClient.mailer != nil {
					if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to restart resolution attempt", err).Error())); err != nil {
						if messageBrokerClient.errorLogger != nil {
							messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
						}
					}
				}
			}
		}
		messageBrokerClient.statusMutex.Unlock()
	case <-stopChannel:
		messageBrokerClient.mutex.Lock()
		for topic := range connection.responsibleAsyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleAsyncTopics, topic)
		}
		for topic := range connection.responsibleSyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleSyncTopics, topic)
		}
		delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
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
			if messageBrokerClient.status != Status.STARTED {
				messageBrokerClient.statusMutex.Unlock()
				return
			}
			err := messageBrokerClient.startResolutionAttempt(topic, isSynctopic, stopChannel)
			if err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to restart resolution attempt for topic \""+topic+"\"", err).Error())
				}
				if messageBrokerClient.mailer != nil {
					if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to restart resolution attempt", err).Error())); err != nil {
						if messageBrokerClient.errorLogger != nil {
							messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
						}
					}
				}
			}
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
					delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
					connection.connection.Close()
				}
			}
			delete(messageBrokerClient.topicResolutions, topic)
		}
	case <-stopChannel:
		return
	}
}
