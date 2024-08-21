package SystemgeClient

func (client *SystemgeClient) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"connection_attempts_failed":      client.GetConnectionAttemptsFailed(),
		"connection_attempts_success":     client.GetConnectionAttemptsSuccess(),
		"bytes_sent":                      client.GetBytesSent(),
		"bytes_received":                  client.GetBytesReceived(),
		"async_messages_sent":             client.GetAsyncMessagesSent(),
		"sync_requests_sent":              client.GetSyncRequestsSent(),
		"sync_success_responses_received": client.GetSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": client.GetSyncFailureResponsesReceived(),
		"no_sync_response_received":       client.GetNoSyncResponseReceived(),
		"async_messages_received":         client.GetAsyncMessagesReceived(),
		"sync_requests_received":          client.GetSyncRequestsReceived(),
		"invalid_messages_received":       client.GetInvalidMessagesReceived(),
		"message_rate_limiter_exceeded":   client.GetMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      client.GetByteRateLimiterExceeded(),
	}
}

func (client *SystemgeClient) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"connection_attempts_failed":      client.RetrieveConnectionAttemptsFailed(),
		"connection_attempts_success":     client.RetrieveConnectionAttemptsSuccess(),
		"bytes_sent":                      client.RetrieveBytesSent(),
		"bytes_received":                  client.RetrieveBytesReceived(),
		"async_messages_sent":             client.RetrieveAsyncMessagesSent(),
		"sync_requests_sent":              client.RetrieveSyncRequestsSent(),
		"sync_success_responses_received": client.RetrieveSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": client.RetrieveSyncFailureResponsesReceived(),
		"no_sync_response_received":       client.RetrieveNoSyncResponseReceived(),
		"async_messages_received":         client.RetrieveAsyncMessagesReceived(),
		"sync_requests_received":          client.RetrieveSyncRequestsReceived(),
		"invalid_messages_received":       client.RetrieveInvalidMessagesReceived(),
		"message_rate_limiter_exceeded":   client.RetrieveMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      client.RetrieveByteRateLimiterExceeded(),
	}
}

func (client *SystemgeClient) GetConnectionAttemptsFailed() uint64 {
	return client.connectionAttemptsFailed.Load()
}
func (client *SystemgeClient) RetrieveConnectionAttemptsFailed() uint64 {
	return client.connectionAttemptsFailed.Swap(0)
}

func (client *SystemgeClient) GetConnectionAttemptsSuccess() uint64 {
	return client.connectionAttemptsSuccess.Load()
}
func (client *SystemgeClient) RetrieveConnectionAttemptsSuccess() uint64 {
	return client.connectionAttemptsSuccess.Swap(0)
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
func (client *SystemgeClient) RetrieveBytesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveBytesSent()
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
func (client *SystemgeClient) RetrieveBytesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveBytesReceived()
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
func (client *SystemgeClient) RetrieveAsyncMessagesSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveAsyncMessagesSent()
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
func (client *SystemgeClient) RetrieveSyncRequestsSent() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveSyncRequestsSent()
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
func (client *SystemgeClient) RetrieveSyncSuccessResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveSyncSuccessResponsesReceived()
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
func (client *SystemgeClient) RetrieveSyncFailureResponsesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveSyncFailureResponsesReceived()
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
func (client *SystemgeClient) RetrieveNoSyncResponseReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveNoSyncResponseReceived()
	}
	return sum
}

func (client *SystemgeClient) GetAsyncMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetAsyncMessagesReceived()
	}
	return sum
}
func (client *SystemgeClient) RetrieveAsyncMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveAsyncMessagesReceived()
	}
	return sum
}

func (client *SystemgeClient) GetSyncRequestsReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.GetSyncRequestsReceived()
	}
	return sum
}
func (client *SystemgeClient) RetrieveSyncRequestsReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveSyncRequestsReceived()
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
func (client *SystemgeClient) RetrieveInvalidMessagesReceived() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveInvalidMessagesReceived()
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
func (client *SystemgeClient) RetrieveMessageRateLimiterExceeded() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveMessageRateLimiterExceeded()
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
func (client *SystemgeClient) RetrieveByteRateLimiterExceeded() uint64 {
	client.statusMutex.RLock()
	client.mutex.Lock()
	defer func() {
		client.mutex.Unlock()
		client.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range client.nameConnections {
		sum += connection.RetrieveByteRateLimiterExceeded()
	}
	return sum
}
