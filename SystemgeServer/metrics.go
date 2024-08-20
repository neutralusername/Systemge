package SystemgeServer

import "github.com/neutralusername/Systemge/Status"

func (server *SystemgeServer) GetConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetConnectionAttempts()
}
func (server *SystemgeServer) RetrieveConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.RetrieveConnectionAttempts()
}

func (server *SystemgeServer) GetFailedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetFailedConnections()
}
func (server *SystemgeServer) RetrieveFailedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.RetrieveFailedConnections()
}

func (server *SystemgeServer) GetRejectedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetRejectedConnections()
}
func (server *SystemgeServer) RetrieveRejectedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.RetrieveRejectedConnections()
}

func (server *SystemgeServer) GetAcceptedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetAcceptedConnections()
}
func (server *SystemgeServer) RetrieveAcceptedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.RetrieveAcceptedConnections()
}

func (server *SystemgeServer) GetBytesSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetBytesSent()
	}
	return sum
}
func (server *SystemgeServer) RetrieveBytesSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveBytesSent()
	}
	return sum
}

func (server *SystemgeServer) GetBytesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetBytesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveBytesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveBytesReceived()
	}
	return sum
}

func (server *SystemgeServer) GetAsyncMessagesSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetAsyncMessagesSent()
	}
	return sum
}
func (server *SystemgeServer) RetrieveAsyncMessagesSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveAsyncMessagesSent()
	}
	return sum
}

func (server *SystemgeServer) GetSyncRequestsSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetSyncRequestsSent()
	}
	return sum
}
func (server *SystemgeServer) RetrieveSyncRequestsSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveSyncRequestsSent()
	}
	return sum
}

func (server *SystemgeServer) GetSyncSuccessResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetSyncSuccessResponsesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveSyncSuccessResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveSyncSuccessResponsesReceived()
	}
	return sum
}

func (server *SystemgeServer) GetSyncFailureResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetSyncFailureResponsesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveSyncFailureResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveSyncFailureResponsesReceived()
	}
	return sum
}

func (server *SystemgeServer) GetNoSyncResponseReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetNoSyncResponseReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveNoSyncResponseReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveNoSyncResponseReceived()
	}
	return sum
}

func (server *SystemgeServer) GetAsyncMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetAsyncMessagesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveAsyncMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveAsyncMessagesReceived()
	}
	return sum
}

func (server *SystemgeServer) GetSyncRequestsReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetSyncRequestsReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveSyncRequestsReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveSyncRequestsReceived()
	}
	return sum
}

func (server *SystemgeServer) GetInvalidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetInvalidMessagesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveInvalidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveInvalidMessagesReceived()
	}
	return sum
}

func (server *SystemgeServer) GetMessageRateLimiterExceeded() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetMessageRateLimiterExceeded()
	}
	return sum
}
func (server *SystemgeServer) RetrieveMessageRateLimiterExceeded() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveMessageRateLimiterExceeded()
	}
	return sum
}

func (server *SystemgeServer) GetByteRateLimiterExceeded() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetByteRateLimiterExceeded()
	}
	return sum
}
func (server *SystemgeServer) RetrieveByteRateLimiterExceeded() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveByteRateLimiterExceeded()
	}
	return sum
}
