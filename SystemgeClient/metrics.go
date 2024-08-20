package SystemgeClient

func (client *SystemgeClient) GetConnectionAttemptsFailed() uint32 {
	return client.connectionAttemptsFailed.Load()
}
func (client *SystemgeClient) RetrieveConnectionAttemptsFailed() uint32 {
	return client.connectionAttemptsFailed.Swap(0)
}

func (client *SystemgeClient) GetConnectionAttemptsSuccess() uint32 {
	return client.connectionAttemptsSuccess.Load()
}
func (client *SystemgeClient) RetrieveConnectionAttemptsSuccess() uint32 {
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
