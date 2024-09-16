package BrokerClient

func (messageBrokerClient *Client) GetAsyncSubscribeTopics() []string {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	topics := []string{}
	for topic := range messageBrokerClient.subscribedAsyncTopics {
		topics = append(topics, topic)
	}
	return topics
}

func (messageBrokerClient *Client) GetSyncSubscribeTopics() []string {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	topics := []string{}
	for topic := range messageBrokerClient.subscribedSyncTopics {
		topics = append(topics, topic)
	}
	return topics
}

func (messageBrokerClient *Client) AddAsyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	messageBrokerClient.subscribedAsyncTopics[topic] = true
	return nil
}

func (messageBrokerClient *Client) AddSyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	messageBrokerClient.subscribedSyncTopics[topic] = true
	return nil
}

func (messageBrokerClient *Client) RemoveAsyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	delete(messageBrokerClient.subscribedAsyncTopics, topic)
	return nil
}

func (messageBrokerClient *Client) RemoveSyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	delete(messageBrokerClient.subscribedSyncTopics, topic)
	return nil
}
