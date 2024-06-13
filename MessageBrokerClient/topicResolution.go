package MessageBrokerClient

import "Systemge/Error"

func (client *Client) addTopicResolution(topic string, serverConnection *serverConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.topicResolutions[topic] != nil {
		return Error.New("Topic resolution already exists", nil)
	}
	client.topicResolutions[topic] = serverConnection
	serverConnection.addTopic(topic)
	return nil
}
func (serverConnection *serverConnection) addTopic(topic string) {
	serverConnection.mapOperationMutex.Lock()
	serverConnection.topics[topic] = true
	serverConnection.mapOperationMutex.Unlock()
}

func (client *Client) getTopicResolution(topic string) *serverConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.topicResolutions[topic]
}
