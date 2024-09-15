package SystemgeServer

import "github.com/neutralusername/Systemge/Status"

func (server *SystemgeServer) CheckMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"systemgeServer": {
			"connection_attempts":             server.CheckConnectionAttempts(),
			"failed_connections":              server.CheckFailedConnections(),
			"rejected_connections":            server.CheckRejectedConnections(),
			"accepted_connections":            server.CheckAcceptedConnections(),
			"bytes_sent":                      server.CheckBytesSent(),
			"bytes_received":                  server.CheckBytesReceived(),
			"async_messages_sent":             server.CheckAsyncMessagesSent(),
			"sync_requests_sent":              server.CheckSyncRequestsSent(),
			"sync_success_responses_received": server.CheckSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": server.CheckSyncFailureResponsesReceived(),
			"no_sync_response_received":       server.CheckNoSyncResponseReceived(),
			"invalid_messages_received":       server.CheckInvalidMessagesReceived(),
			"invalid_sync_responses_received": server.CheckInvalidSyncResponsesReceived(),
			"valid_messages_received":         server.CheckValidMessagesReceived(),
			"message_rate_limiter_exceeded":   server.CheckMessageRateLimiterExceeded(),
			"byte_rate_limiter_exceeded":      server.CheckByteRateLimiterExceeded(),
			"connection_count":                uint64(server.GetConnectionCount()),
		},
	}
}

func (server *SystemgeServer) GetMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"systemgeServer": {
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
			"invalid_sync_responses_received": server.GetInvalidSyncResponsesReceived(),
			"valid_messages_received":         server.GetValidMessagesReceived(),
			"message_rate_limiter_exceeded":   server.GetMessageRateLimiterExceeded(),
			"byte_rate_limiter_exceeded":      server.GetByteRateLimiterExceeded(),
			"connection_count":                uint64(server.GetConnectionCount()),
		},
	}
}

func (server *SystemgeServer) CheckConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.CheckConnectionAttempts()
}
func (server *SystemgeServer) GetConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetConnectionAttempts()
}

func (server *SystemgeServer) CheckFailedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.CheckFailedConnections()
}
func (server *SystemgeServer) GetFailedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetFailedConnections()
}

func (server *SystemgeServer) CheckRejectedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.CheckRejectedConnections()
}
func (server *SystemgeServer) GetRejectedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetRejectedConnections()
}

func (server *SystemgeServer) CheckAcceptedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.CheckAcceptedConnections()
}
func (server *SystemgeServer) GetAcceptedConnections() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.STARTED {
		return 0
	}

	return server.listener.GetAcceptedConnections()
}

func (server *SystemgeServer) CheckBytesSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckBytesSent()
	}
	return sum
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

func (server *SystemgeServer) CheckBytesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckBytesReceived()
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

func (server *SystemgeServer) CheckAsyncMessagesSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckAsyncMessagesSent()
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

func (server *SystemgeServer) CheckSyncRequestsSent() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckSyncRequestsSent()
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

func (server *SystemgeServer) CheckSyncSuccessResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckSyncSuccessResponsesReceived()
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

func (server *SystemgeServer) CheckSyncFailureResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckSyncFailureResponsesReceived()
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

func (server *SystemgeServer) CheckNoSyncResponseReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckNoSyncResponseReceived()
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

func (server *SystemgeServer) CheckInvalidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckInvalidMessagesReceived()
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

func (server *SystemgeServer) CheckInvalidSyncResponsesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckInvalidSyncResponsesReceived()
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

func (server *SystemgeServer) CheckValidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckValidMessagesReceived()
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

func (server *SystemgeServer) CheckMessageRateLimiterExceeded() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckMessageRateLimiterExceeded()
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

func (server *SystemgeServer) CheckByteRateLimiterExceeded() uint64 {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()

	sum := uint64(0)
	for _, connection := range server.clients {
		sum += connection.CheckByteRateLimiterExceeded()
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
