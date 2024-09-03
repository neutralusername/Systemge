package BrokerClient

func (messageBrokerClient *Client) getTopicResolutions(topic string) ([]*connection, error) {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()

	if messageBrokerClient.topicResolutions[topic] == nil {
		// resolve topic
	}
	connectionList := []*connection{}
	for _, connection := range messageBrokerClient.topicResolutions[topic] {
		connectionList = append(connectionList, connection)
	}
	return connectionList, nil
}
