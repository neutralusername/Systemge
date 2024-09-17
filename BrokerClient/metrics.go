package BrokerClient

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (messageBrokerClient *Client) CheckMetrics() map[string]*Metrics.Metrics {
	metricsTypes := Metrics.NewMetricsTypes()
	messageBrokerClient.mutex.Lock()
	metricsTypes.AddMetrics("broker_client", Metrics.New(
		map[string]uint64{
			"ongoing_topic_resolutions": uint64(len(messageBrokerClient.ongoingTopicResolutions)),
			"broker_connections":        uint64(len(messageBrokerClient.brokerConnections)),
			"topic_resolutions":         uint64(len(messageBrokerClient.topicResolutions)),
			"resolution_attempts":       messageBrokerClient.CheckResolutionAttempts(),
			"async_messages_sent":       messageBrokerClient.CheckAsyncMessagesSent(),
			"sync_requests_sent":        messageBrokerClient.CheckSyncRequestsSent(),
			"sync_responses_received":   messageBrokerClient.CheckSyncResponsesReceived(),
		},
	))
	messageBrokerClient.mutex.Unlock()
	keyValuePairs := map[string]uint64{}
	for _, connection := range messageBrokerClient.brokerConnections {
		for _, metricsMap := range connection.connection.CheckMetrics() {
			for key, value := range metricsMap.KeyValuePairs {
				keyValuePairs[key] += value
			}
		}
	}
	metricsTypes.AddMetrics("broker_connections", Metrics.New(keyValuePairs))
	return metricsTypes
}
func (messageBrokerClient *Client) GetMetrics() map[string]*Metrics.Metrics {
	metricsTypes := Metrics.NewMetricsTypes()
	messageBrokerClient.mutex.Lock()
	metricsTypes.AddMetrics("broker_client", Metrics.New(
		map[string]uint64{
			"ongoing_topic_resolutions": uint64(len(messageBrokerClient.ongoingTopicResolutions)),
			"broker_connections":        uint64(len(messageBrokerClient.brokerConnections)),
			"topic_resolutions":         uint64(len(messageBrokerClient.topicResolutions)),
			"resolution_attempts":       messageBrokerClient.GetResolutionAttempts(),
			"async_messages_sent":       messageBrokerClient.GetAsyncMessagesSent(),
			"sync_requests_sent":        messageBrokerClient.GetSyncRequestsSent(),
			"sync_responses_received":   messageBrokerClient.GetSyncResponsesReceived(),
		},
	))
	messageBrokerClient.mutex.Unlock()
	keyValuePairs := map[string]uint64{}
	for _, connection := range messageBrokerClient.brokerConnections {
		for _, metricsMap := range connection.connection.GetMetrics() {
			for key, value := range metricsMap.KeyValuePairs {
				keyValuePairs[key] += value
			}
		}
	}
	metricsTypes.AddMetrics("broker_connections", Metrics.New(keyValuePairs))
	return metricsTypes
}

func (messageBrokerClient *Client) CheckAsyncMessagesSent() uint64 {
	return messageBrokerClient.asyncMessagesSent.Load()
}
func (messageBrokerClient *Client) GetAsyncMessagesSent() uint64 {
	return messageBrokerClient.asyncMessagesSent.Swap(0)
}

func (messageBrokerClient *Client) CheckSyncRequestsSent() uint64 {
	return messageBrokerClient.syncRequestsSent.Load()
}
func (messageBrokerClient *Client) GetSyncRequestsSent() uint64 {
	return messageBrokerClient.syncRequestsSent.Swap(0)
}

func (messageBrokerClient *Client) CheckSyncResponsesReceived() uint64 {
	return messageBrokerClient.syncResponsesReceived.Load()
}
func (messageBrokerClient *Client) GetSyncResponsesReceived() uint64 {
	return messageBrokerClient.syncResponsesReceived.Swap(0)
}

func (messageBrokerClient *Client) CheckResolutionAttempts() uint64 {
	return messageBrokerClient.resolutionAttempts.Load()
}
func (messageBrokerClient *Client) GetResolutionAttempts() uint64 {
	return messageBrokerClient.resolutionAttempts.Swap(0)
}
