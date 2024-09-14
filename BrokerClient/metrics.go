package BrokerClient

func (messageBrokerClient *Client) GetMetrics_() map[string]map[string]uint64 {
	metrics := map[string]map[string]uint64{}
	metrics["broker_client"] = map[string]uint64{}
	metrics["broker_client"]["ongoing_topic_resolutions"] = uint64(len(messageBrokerClient.ongoingTopicResolutions))
	metrics["broker_client"]["broker_connections"] = uint64(len(messageBrokerClient.brokerConnections))
	metrics["broker_client"]["topic_resolutions"] = uint64(len(messageBrokerClient.topicResolutions))
	metrics["broker_client"]["resolution_attempts"] = messageBrokerClient.GetResolutionAttempts()
	metrics["broker_client"]["async_messages_sent"] = messageBrokerClient.GetAsyncMessagesSent()
	metrics["broker_client"]["sync_requests_sent"] = messageBrokerClient.GetSyncRequestsSent()
	metrics["broker_client"]["sync_responses_received"] = messageBrokerClient.GetSyncResponsesReceived()
	metrics["broker_connections"] = map[string]uint64{}
	for _, connection := range messageBrokerClient.brokerConnections {
		for _, metricsMap := range connection.connection.GetMetrics_() {
			for key, value := range metricsMap {
				metrics["broker_connections"][key] += value
			}
		}
	}
	return metrics
}
func (messageBrokerClient *Client) GetMetrics() map[string]map[string]uint64 {
	metrics := map[string]map[string]uint64{}
	metrics["broker_client"] = map[string]uint64{}
	metrics["broker_client"]["ongoing_topic_resolutions"] = uint64(len(messageBrokerClient.ongoingTopicResolutions))
	metrics["broker_client"]["broker_connections"] = uint64(len(messageBrokerClient.brokerConnections))
	metrics["broker_client"]["topic_resolutions"] = uint64(len(messageBrokerClient.topicResolutions))
	metrics["broker_client"]["resolution_attempts"] = messageBrokerClient.RetrieveResolutionAttempts()
	metrics["broker_client"]["async_messages_sent"] = messageBrokerClient.RetrieveAsyncMessagesSent()
	metrics["broker_client"]["sync_requests_sent"] = messageBrokerClient.RetrieveSyncRequestsSent()
	metrics["broker_client"]["sync_responses_received"] = messageBrokerClient.RetrieveSyncResponsesReceived()
	metrics["broker_connections"] = map[string]uint64{}
	for _, connection := range messageBrokerClient.brokerConnections {
		for _, metricsMap := range connection.connection.GetMetrics() {
			for key, value := range metricsMap {
				metrics["broker_connections"][key] += value
			}
		}
	}
	return metrics
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
