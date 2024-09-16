package BrokerClient

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

type resolutionAttempt struct {
	topic       string
	isSyncTopic bool
	ongoing     chan bool
	connections map[string]*connection
}

func (messageBrokerClient *Client) startResolutionAttempt(topic string, syncTopic bool, stopChannel chan bool, subscribe bool) (*resolutionAttempt, error) {
	if stopChannel != messageBrokerClient.stopChannel {
		return nil, Error.New("Aborted because resolution attempt belongs to outdated session", nil)
	}
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()

	if resolutionAttempt := messageBrokerClient.ongoingTopicResolutions[topic]; resolutionAttempt != nil {
		return resolutionAttempt, nil
	}
	resolutionAttempt := &resolutionAttempt{
		ongoing:     make(chan bool),
		topic:       topic,
		isSyncTopic: syncTopic,
	}
	messageBrokerClient.ongoingTopicResolutions[topic] = resolutionAttempt
	messageBrokerClient.waitGroup.Add(1)
	messageBrokerClient.resolutionAttempts.Add(1)
	go func() {
		messageBrokerClient.resolutionAttempt(resolutionAttempt, stopChannel, subscribe)
	}()
	return resolutionAttempt, nil
}

func (messageBrokerClient *Client) resolutionAttempt(resolutionAttempt *resolutionAttempt, stopChannel chan bool, subscribe bool) {
	endpoints := messageBrokerClient.resolveBrokerEndpoints(resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
	attempts := uint32(0)
	for subscribe &&
		len(endpoints) == 0 &&
		stopChannel == messageBrokerClient.stopChannel &&
		(messageBrokerClient.config.ResolutionAttemptMaxAttempts == 0 || attempts < messageBrokerClient.config.ResolutionAttemptMaxAttempts) {

		time.Sleep(time.Duration(messageBrokerClient.config.ResolutionAttemptRetryIntervalMs) * time.Millisecond)
		endpoints = messageBrokerClient.resolveBrokerEndpoints(resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
		attempts++
	}

	connections := map[string]*connection{}
	for _, endpoint := range endpoints {
		conn, err := messageBrokerClient.getBrokerConnection(endpoint, stopChannel)
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to get connection to resolved endpoint \""+endpoint.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to get connection to resolved endpoint \""+endpoint.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		}
		messageBrokerClient.mutex.Lock()
		if resolutionAttempt.isSyncTopic {
			conn.responsibleSyncTopics[resolutionAttempt.topic] = true
		} else {
			conn.responsibleAsyncTopics[resolutionAttempt.topic] = true
		}
		connections[getEndpointString(endpoint)] = conn
		messageBrokerClient.mutex.Unlock()

		if subscribe {
			messageBrokerClient.subscribeToTopic(conn, resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
		}
	}

	messageBrokerClient.mutex.Lock()
	for endpointString, existingConnection := range messageBrokerClient.topicResolutions[resolutionAttempt.topic] {
		if _, ok := connections[endpointString]; !ok {
			if resolutionAttempt.isSyncTopic {
				delete(existingConnection.responsibleSyncTopics, resolutionAttempt.topic)
			} else {
				delete(existingConnection.responsibleAsyncTopics, resolutionAttempt.topic)
			}
			if len(existingConnection.responsibleAsyncTopics) == 0 && len(existingConnection.responsibleSyncTopics) == 0 {
				delete(messageBrokerClient.brokerConnections, endpointString)
				existingConnection.connection.Close()
			}
		}
	}
	if len(connections) > 0 {
		messageBrokerClient.topicResolutions[resolutionAttempt.topic] = connections
		go messageBrokerClient.handleTopicResolutionLifetime(resolutionAttempt.topic, resolutionAttempt.isSyncTopic, stopChannel)
	}
	resolutionAttempt.connections = connections
	delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
	close(resolutionAttempt.ongoing)
	messageBrokerClient.mutex.Unlock()

	messageBrokerClient.waitGroup.Done()
}
