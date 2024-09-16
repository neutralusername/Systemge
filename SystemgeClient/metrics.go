package SystemgeClient

import (
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

func (client *SystemgeClient) CheckMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"systemge_client": {
			KeyValuePairs: map[string]uint64{
				"connection_attempts_failed":      client.CheckConnectionAttemptsFailed(),
				"connection_attempts_success":     client.CheckConnectionAttemptsSuccess(),
				"bytes_sent":                      client.CheckBytesSent(),
				"bytes_received":                  client.CheckBytesReceived(),
				"async_messages_sent":             client.CheckAsyncMessagesSent(),
				"sync_requests_sent":              client.CheckSyncRequestsSent(),
				"sync_success_responses_received": client.CheckSyncSuccessResponsesReceived(),
				"sync_failure_responses_received": client.CheckSyncFailureResponsesReceived(),
				"no_sync_response_received":       client.CheckNoSyncResponseReceived(),
				"invalid_messages_received":       client.CheckInvalidMessagesReceived(),
				"invalid_sync_responses_received": client.CheckInvalidSyncResponsesReceived(),
				"valid_messages_received":         client.CheckValidMessagesReceived(),
				"message_rate_limiter_exceeded":   client.CheckMessageRateLimiterExceeded(),
				"byte_rate_limiter_exceeded":      client.CheckByteRateLimiterExceeded(),
			},
			Time: time.Now(),
		},
	}
}

func (client *SystemgeClient) GetMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"systemge_client": {
			KeyValuePairs: map[string]uint64{
				"connection_attempts_failed":      client.GetConnectionAttemptsFailed(),
				"connection_attempts_success":     client.GetConnectionAttemptsSuccess(),
				"bytes_sent":                      client.GetBytesSent(),
				"bytes_received":                  client.GetBytesReceived(),
				"async_messages_sent":             client.GetAsyncMessagesSent(),
				"sync_requests_sent":              client.GetSyncRequestsSent(),
				"sync_success_responses_received": client.GetSyncSuccessResponsesReceived(),
				"sync_failure_responses_received": client.GetSyncFailureResponsesReceived(),
				"no_sync_response_received":       client.GetNoSyncResponseReceived(),
				"invalid_messages_received":       client.GetInvalidMessagesReceived(),
				"invalid_sync_responses_received": client.GetInvalidSyncResponsesReceived(),
				"valid_messages_received":         client.GetValidMessagesReceived(),
				"message_rate_limiter_exceeded":   client.GetMessageRateLimiterExceeded(),
				"byte_rate_limiter_exceeded":      client.GetByteRateLimiterExceeded(),
			},
			Time: time.Now(),
		},
	}
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

func (client *SystemgeClient) CheckInvalidSyncResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckInvalidSyncResponsesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetInvalidSyncResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetInvalidSyncResponsesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckValidMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckValidMessagesReceived()
	}
	return sum
}
func (client *SystemgeClient) GetValidMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetValidMessagesReceived()
	}
	return sum
}

func (client *SystemgeClient) CheckMessageRateLimiterExceeded() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckMessageRateLimiterExceeded()
	}
	return sum
}
func (client *SystemgeClient) GetMessageRateLimiterExceeded() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetMessageRateLimiterExceeded()
	}
	return sum
}

func (client *SystemgeClient) CheckByteRateLimiterExceeded() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.CheckByteRateLimiterExceeded()
	}
	return sum
}
func (client *SystemgeClient) GetByteRateLimiterExceeded() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetByteRateLimiterExceeded()
	}
	return sum
}
