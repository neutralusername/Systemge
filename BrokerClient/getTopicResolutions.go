package BrokerClient

import (
	"github.com/neutralusername/Systemge/Error"
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
			return nil, Error.New("Not started", nil)
		}
		resolutionAttempt, err := messageBrokerClient.startResolutionAttempt(topic, isSyncTopic, messageBrokerClient.stopChannel)
		messageBrokerClient.statusMutex.Unlock()
		if err != nil {
			return nil, err
		}
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
