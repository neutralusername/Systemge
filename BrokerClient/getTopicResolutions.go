package BrokerClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (messageBrokerClient *Client) getTopicResolutions(topic string) ([]SystemgeConnection.SystemgeConnection, error) {
	messageBrokerClient.mutex.Lock()
	topicResolutions := messageBrokerClient.topicResolutions[topic]
	messageBrokerClient.mutex.Unlock()

	if topicResolutions == nil {
		messageBrokerClient.statusMutex.Lock()
		if messageBrokerClient.status != Status.STARTED {
			messageBrokerClient.statusMutex.Unlock()
			return nil, Error.New("Not started", nil)
		}
		attempt, err := messageBrokerClient.startResolutionAttempt(topic)
		if err != nil {
			messageBrokerClient.statusMutex.Unlock()
			return nil, err
		}
		messageBrokerClient.statusMutex.Unlock()
		<-attempt.ongoing

		return connectionList, nil
	}
	systemgeConnections := make([]SystemgeConnection.SystemgeConnection, 0, len(topicResolutions))
	for _, connection := range topicResolutions {
		systemgeConnections = append(systemgeConnections, connection)
	}
	return systemgeConnections, nil
}
