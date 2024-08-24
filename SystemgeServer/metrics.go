package SystemgeServer

import "github.com/neutralusername/Systemge/Status"

func (server *SystemgeServer) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"connection_attempts":             server.GetConnectionAttempts(),
		"failed_connections":              server.GetFailedConnections(),
		"rejected_connections":            server.GetRejectedConnections(),
		"accepted_connections":            server.GetAcceptedConnections(),
		"bytes_sent":                      server.GetBytesSent(),
		"bytes_received":                  server.GetBytesReceived(),
		"async_messages_sent":             server.GetAsyncMessagesSent(),
		"sync_requests_sent":              server.GetSyncRequestsSent(),
		"sync_success_responses_received": server.GetSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": server.GetSyncFailureResponsesReceived(),
		"no_sync_response_received":       server.GetNoSyncResponseReceived(),
		"invalid_messages_received":       server.GetInvalidMessagesReceived(),
		"message_rate_limiter_exceeded":   server.GetMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      server.GetByteRateLimiterExceeded(),
	}
}

func (server *SystemgeServer) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"connection_attempts":             server.RetrieveConnectionAttempts(),
		"failed_connections":              server.RetrieveFailedConnections(),
		"rejected_connections":            server.RetrieveRejectedConnections(),
		"accepted_connections":            server.RetrieveAcceptedConnections(),
		"bytes_sent":                      server.RetrieveBytesSent(),
		"bytes_received":                  server.RetrieveBytesReceived(),
		"async_messages_sent":             server.RetrieveAsyncMessagesSent(),
		"sync_requests_sent":              server.RetrieveSyncRequestsSent(),
		"sync_success_responses_received": server.RetrieveSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": server.RetrieveSyncFailureResponsesReceived(),
		"no_sync_response_received":       server.RetrieveNoSyncResponseReceived(),
		"invalid_messages_received":       server.RetrieveInvalidMessagesReceived(),
		"message_rate_limiter_exceeded":   server.RetrieveMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      server.RetrieveByteRateLimiterExceeded(),
	}
}

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

func (server *SystemgeServer) GetInvalidSyncResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetInvalidSyncResponsesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveInvalidSyncResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveInvalidSyncResponsesReceived()
	}
	return sum
}

func (server *SystemgeServer) GetValidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.GetValidMessagesReceived()
	}
	return sum
}
func (server *SystemgeServer) RetrieveValidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.RetrieveValidMessagesReceived()
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
