package BrokerClient

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (messageBrokerClient *Client) getTopicResolutions(topic string, isSyncTopic bool) ([]*connection, error) {
	messageBrokerClient.mutex.Lock()
	topicResolutions := messageBrokerClient.topicResolutions[topic]
	messageBrokerClient.mutex.Unlock()

	if topicResolutions == nil {
		messageBrokerClient.statusMutex.Lock()
		if messageBrokerClient.status != Status.STARTED {
			messageBrokerClient.statusMutex.Unlock()
			return nil, Event.New("Not started", nil)
		}
		var resolutionAttempt *resolutionAttempt
		if isSyncTopic {
			attempt, err := messageBrokerClient.startResolutionAttempt(topic, isSyncTopic, messageBrokerClient.stopChannel)
			resolutionAttempt = attempt
			if err != nil {
				messageBrokerClient.statusMutex.Unlock()
				return nil, err
			}
		} else {
			attempt, err := messageBrokerClient.startResolutionAttempt(topic, isSyncTopic, messageBrokerClient.stopChannel)
			resolutionAttempt = attempt
			if err != nil {
				messageBrokerClient.statusMutex.Unlock()
				return nil, err
			}
		}
		messageBrokerClient.statusMutex.Unlock()
		<-resolutionAttempt.ongoing
		connectionList := []*connection{}
		for _, connection := range resolutionAttempt.connections {
			connectionList = append(connectionList, connection)
		}
		return connectionList, nil
	}
	connectionList := []*connection{}
	for _, connection := range topicResolutions {
		connectionList = append(connectionList, connection)
	}
	return connectionList, nil
}
