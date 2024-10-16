package Server

import (
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *Server) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("systemgeServer_connectionAttempts", Metrics.New(
		map[string]uint64{
			"connection_attempts":          server.CheckConnectionAttempts(),
			"failed_connection_attempts":   server.CheckFailedConnectionAttempts(),
			"rejected_connection_attempts": server.CheckRejectedConnectionAttempts(),
			"accepted_connection_attempts": server.CheckAcceptedConnectionAttempts(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     server.CheckBytesSent(),
			"bytes_received": server.CheckBytesReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_messagesSent", Metrics.New(
		map[string]uint64{
			"async_messages_sent": server.CheckAsyncMessagesSent(),
			"sync_requests_sent":  server.CheckSyncRequestsSent(),
			"sync_responses_sent": server.CheckSyncResponsesSent(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			"sync_success_responses_received": server.CheckSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": server.CheckSyncFailureResponsesReceived(),
			"no_sync_response_received":       server.CheckNoSyncResponseReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_messagesReceived", Metrics.New(
		map[string]uint64{
			"invalid_messages_received": server.CheckInvalidMessagesReceived(),
			"messages_received":         server.CheckMessagesReceived(),
			"rejected_messages":         server.CheckRejectedMessages(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_sessionManager", Metrics.New(
		map[string]uint64{
			// todo (sessions count and identities count)
		},
	))
	return metricsTypes
}

func (server *Server) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("systemgeServer_connectionAttempts", Metrics.New(
		map[string]uint64{
			"connection_attempts":          server.GetConnectionAttempts(),
			"failed_connection_attempts":   server.GetFailedConnectionAttempts(),
			"rejected_connection_attempts": server.GetRejectedConnectionAttempts(),
			"accepted_connection_attempts": server.GetAcceptedConnectionAttempts(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     server.GetBytesSent(),
			"bytes_received": server.GetBytesReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_messagesSent", Metrics.New(
		map[string]uint64{
			"async_messages_sent": server.GetAsyncMessagesSent(),
			"sync_requests_sent":  server.GetSyncRequestsSent(),
			"sync_responses_sent": server.GetSyncResponsesSent(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			"sync_success_responses_received": server.GetSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": server.GetSyncFailureResponsesReceived(),
			"no_sync_response_received":       server.GetNoSyncResponseReceived(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_messagesReceived", Metrics.New(
		map[string]uint64{
			"invalid_messages_received": server.GetInvalidMessagesReceived(),
			"messages_received":         server.GetMessagesReceived(),
			"rejected_messages":         server.GetRejectedMessages(),
		},
	))
	metricsTypes.AddMetrics("systemgeServer_sessionManager", Metrics.New(
		map[string]uint64{
			// todo (sessions count and identities count)
		},
	))
	return metricsTypes
}

func (server *Server) CheckConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckConnectionAttempts()
}
func (server *Server) GetConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetConnectionAttempts()
}

func (server *Server) CheckFailedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckFailedConnectionAttempts()
}
func (server *Server) GetFailedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetFailedConnectionAttempts()
}

func (server *Server) CheckRejectedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckRejectedConnectionAttempts()
}
func (server *Server) GetRejectedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetRejectedConnectionAttempts()
}

func (server *Server) CheckAcceptedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckAcceptedConnectionAttempts()
}
func (server *Server) GetAcceptedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetAcceptedConnectionAttempts()
}

func (server *Server) CheckBytesSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckBytesSent()
		}
	}
	return sum
}
func (server *Server) GetBytesSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetBytesSent()
		}
	}
	return sum
}

func (server *Server) CheckBytesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckBytesReceived()
		}
	}
	return sum
}
func (server *Server) GetBytesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetBytesReceived()
		}
	}
	return sum
}

func (server *Server) CheckAsyncMessagesSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckAsyncMessagesSent()
		}
	}
	return sum
}
func (server *Server) GetAsyncMessagesSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetAsyncMessagesSent()
		}
	}
	return sum
}

func (server *Server) CheckSyncRequestsSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckSyncRequestsSent()
		}
	}
	return sum
}
func (server *Server) GetSyncRequestsSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetSyncRequestsSent()
		}
	}
	return sum
}

func (server *Server) CheckSyncSuccessResponsesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckSyncSuccessResponsesReceived()
		}
	}
	return sum
}
func (server *Server) GetSyncSuccessResponsesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetSyncSuccessResponsesReceived()
		}
	}
	return sum
}

func (server *Server) CheckSyncFailureResponsesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckSyncFailureResponsesReceived()
		}
	}
	return sum
}
func (server *Server) GetSyncFailureResponsesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetSyncFailureResponsesReceived()
		}
	}
	return sum
}

func (server *Server) CheckNoSyncResponseReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckNoSyncResponseReceived()
		}
	}
	return sum
}
func (server *Server) GetNoSyncResponseReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetNoSyncResponseReceived()
		}
	}
	return sum
}

func (server *Server) CheckInvalidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckInvalidMessagesReceived()
		}
	}
	return sum
}
func (server *Server) GetInvalidMessagesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetInvalidMessagesReceived()
		}
	}
	return sum
}

func (server *Server) CheckSyncResponsesSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckSyncResponsesSent()
		}
	}
	return sum
}
func (server *Server) GetSyncResponsesSent() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetSyncResponsesSent()
		}
	}
	return sum
}

func (server *Server) CheckMessagesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckMessagesReceived()
		}
	}
	return sum
}
func (server *Server) GetMessagesReceived() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetMessagesReceived()
		}
	}
	return sum
}

func (server *Server) CheckRejectedMessages() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).CheckRejectedMessages()
		}
	}
	return sum
}
func (server *Server) GetRejectedMessages() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	sum := uint64(0)
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			sum += connection.(SystemgeConnection.SystemgeConnection).GetRejectedMessages()
		}
	}
	return sum
}
