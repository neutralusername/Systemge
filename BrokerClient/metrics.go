package BrokerClient

func (messageBrokerClient *Client) GetMetrics() map[string]uint64 {
	m := map[string]uint64{}
	messageBrokerClient.mutex.Lock()
	m["ongoing_topic_resolutions"] = uint64(len(messageBrokerClient.ongoingTopicResolutions))
	m["broker_connections"] = uint64(len(messageBrokerClient.brokerConnections))
	m["topic_resolutions"] = uint64(len(messageBrokerClient.topicResolutions))
	m["resolution_attempts"] = messageBrokerClient.GetResolutionAttempts()
	m["async_messages_sent"] = messageBrokerClient.GetAsyncMessagesSent()
	m["sync_requests_sent"] = messageBrokerClient.GetSyncRequestsSent()
	m["sync_responses_received"] = messageBrokerClient.GetSyncResponsesReceived()
	for _, connection := range messageBrokerClient.brokerConnections {
		metrics := connection.connection.GetMetrics()
		for key, value := range metrics {
			m[key] += value
		}
	}
	messageBrokerClient.mutex.Unlock()
	return m
}
func (messageBrokerClient *Client) RetrieveMetrics() map[string]uint64 {
	m := map[string]uint64{}
	messageBrokerClient.mutex.Lock()
	m["ongoing_topic_resolutions"] = uint64(len(messageBrokerClient.ongoingTopicResolutions))
	m["broker_connections"] = uint64(len(messageBrokerClient.brokerConnections))
	m["topic_resolutions"] = uint64(len(messageBrokerClient.topicResolutions))
	m["resolution_attempts"] = messageBrokerClient.RetrieveResolutionAttempts()
	m["async_messages_sent"] = messageBrokerClient.RetrieveAsyncMessagesSent()
	m["sync_requests_sent"] = messageBrokerClient.RetrieveSyncRequestsSent()
	m["sync_responses_received"] = messageBrokerClient.RetrieveSyncResponsesReceived()
	for _, connection := range messageBrokerClient.brokerConnections {
		metrics := connection.connection.RetrieveMetrics()
		for key, value := range metrics {
			m[key] += value
		}
	}
	messageBrokerClient.mutex.Unlock()
	return m
}

func (messageBrokerClient *Client) GetAsyncMessagesSent() uint64 {
	return messageBrokerClient.asyncMessagesSent.Load()
}
func (messageBrokerClient *Client) RetrieveAsyncMessagesSent() uint64 {
	return messageBrokerClient.asyncMessagesSent.Swap(0)
}

func (messageBrokerClient *Client) GetSyncRequestsSent() uint64 {
	return messageBrokerClient.syncRequestsSent.Load()
}
func (messageBrokerClient *Client) RetrieveSyncRequestsSent() uint64 {
	return messageBrokerClient.syncRequestsSent.Swap(0)
}

func (messageBrokerClient *Client) GetSyncResponsesReceived() uint64 {
	return messageBrokerClient.syncResponsesReceived.Load()
}
func (messageBrokerClient *Client) RetrieveSyncResponsesReceived() uint64 {
	return messageBrokerClient.syncResponsesReceived.Swap(0)
}

func (messageBrokerClient *Client) GetResolutionAttempts() uint64 {
	return messageBrokerClient.resolutionAttempts.Load()
}
func (messageBrokerClient *Client) RetrieveResolutionAttempts() uint64 {
	return messageBrokerClient.resolutionAttempts.Swap(0)
}
