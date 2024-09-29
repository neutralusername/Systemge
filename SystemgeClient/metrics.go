package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (client *SystemgeClient) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("systemgeClient_connectionAttempts", Metrics.New(
		map[string]uint64{
			"connection_attempts_failed":  client.CheckConnectionAttemptsFailed(),
			"connection_attempts_success": client.CheckConnectionAttemptsSuccess(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     client.CheckBytesSent(),
			"bytes_received": client.CheckBytesReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_messagesSent", Metrics.New(
		map[string]uint64{
			"async_messages_sent": client.CheckAsyncMessagesSent(),
			"sync_requests_sent":  client.CheckSyncRequestsSent(),
			"sync_responses_sent": client.CheckSyncResponsesSent(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			"sync_success_responses_received": client.CheckSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": client.CheckSyncFailureResponsesReceived(),
			"no_sync_response_received":       client.CheckNoSyncResponseReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_messagesReceived", Metrics.New(
		map[string]uint64{
			"invalid_messages_received": client.CheckInvalidMessagesReceived(),
			"messages_received":         client.CheckMessagesReceived(),
			"rejected_messages":         client.CheckRejectedMessages(),
		},
	))
	return metricsTypes
}

func (client *SystemgeClient) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("systemgeClient_connectionAttempts", Metrics.New(
		map[string]uint64{
			"connection_attempts_failed":  client.GetConnectionAttemptsFailed(),
			"connection_attempts_success": client.GetConnectionAttemptsSuccess(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     client.GetBytesSent(),
			"bytes_received": client.GetBytesReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_messagesSent", Metrics.New(
		map[string]uint64{
			"async_messages_sent": client.GetAsyncMessagesSent(),
			"sync_requests_sent":  client.GetSyncRequestsSent(),
			"sync_responses_sent": client.GetSyncResponsesSent(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			"sync_success_responses_received": client.GetSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": client.GetSyncFailureResponsesReceived(),
			"no_sync_response_received":       client.GetNoSyncResponseReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeClient_messagesReceived", Metrics.New(
		map[string]uint64{
			"invalid_messages_received": client.GetInvalidMessagesReceived(),
			"messages_received":         client.GetMessagesReceived(),
			"rejected_messages":         client.GetRejectedMessages(),
		},
	))
	return metricsTypes
}

func (client *SystemgeClient) CheckConnectionAttemptsFailed() uint64 {
	return client.connectionAttemptsFailed.Load()
}
func (client *SystemgeClient) GetConnectionAttemptsFailed() uint64 {
	return client.connectionAttemptsFailed.Swap(0)
}

func (client *SystemgeClient) CheckConnectionAttemptsSuccess() uint64 {
	return client.connectionAttemptsSuccess.Load()
}
func (client *SystemgeClient) GetConnectionAttemptsSuccess() uint64 {
	return client.connectionAttemptsSuccess.Swap(0)
}

func (client *SystemgeClient) CheckBytesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckBytesSent()
	}
	return sum
}
func (client *SystemgeClient) GetBytesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetBytesSent()
	}
	return sum
}

func (client *SystemgeClient) CheckBytesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckBytesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetBytesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetBytesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckAsyncMessagesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckAsyncMessagesSent()
	}
	return sum
}
func (client *SystemgeClient) GetAsyncMessagesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetAsyncMessagesSent()
	}
	return sum
}

func (client *SystemgeClient) CheckSyncRequestsSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckSyncRequestsSent()
	}
	return sum
}
func (client *SystemgeClient) GetSyncRequestsSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetSyncRequestsSent()
	}
	return sum
}

func (client *SystemgeClient) CheckSyncResponsesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckSyncResponsesSent()
	}
	return sum
}
func (client *SystemgeClient) GetSyncResponsesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetSyncResponsesSent()
	}
	return sum
}

func (client *SystemgeClient) CheckSyncSuccessResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckSyncSuccessResponsesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetSyncSuccessResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetSyncSuccessResponsesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckSyncFailureResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckSyncFailureResponsesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetSyncFailureResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetSyncFailureResponsesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckNoSyncResponseReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckNoSyncResponseReceived()
	}
	return sum
}
func (client *SystemgeClient) GetNoSyncResponseReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetNoSyncResponseReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckInvalidMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckInvalidMessagesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetInvalidMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetInvalidMessagesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckMessagesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetMessagesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckRejectedMessages() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckRejectedMessages()
	}
	return sum
}
func (client *SystemgeClient) GetRejectedMessages() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetRejectedMessages()
	}
	return sum
}
