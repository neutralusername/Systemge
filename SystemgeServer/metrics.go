package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) CheckMetrics() Metrics.MetricsTypes {
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

func (server *SystemgeServer) GetMetrics() Metrics.MetricsTypes {
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

func (server *SystemgeServer) CheckConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckConnectionAttempts()
}
func (server *SystemgeServer) GetConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetConnectionAttempts()
}

func (server *SystemgeServer) CheckFailedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckFailedConnectionAttempts()
}
func (server *SystemgeServer) GetFailedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetFailedConnectionAttempts()
}

func (server *SystemgeServer) CheckRejectedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckRejectedConnectionAttempts()
}
func (server *SystemgeServer) GetRejectedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetRejectedConnectionAttempts()
}

func (server *SystemgeServer) CheckAcceptedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.CheckAcceptedConnectionAttempts()
}
func (server *SystemgeServer) GetAcceptedConnectionAttempts() uint64 {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()
	if server.status != Status.Started {
		return 0
	}

	return server.listener.GetAcceptedConnectionAttempts()
}

func (server *SystemgeServer) CheckBytesSent() uint64 {
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
func (server *SystemgeServer) GetBytesSent() uint64 {
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

func (server *SystemgeServer) CheckBytesReceived() uint64 {
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
func (server *SystemgeServer) GetBytesReceived() uint64 {
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

func (server *SystemgeServer) CheckAsyncMessagesSent() uint64 {
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
func (server *SystemgeServer) GetAsyncMessagesSent() uint64 {
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

func (server *SystemgeServer) CheckSyncRequestsSent() uint64 {
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
func (server *SystemgeServer) GetSyncRequestsSent() uint64 {
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

func (server *SystemgeServer) CheckSyncSuccessResponsesReceived() uint64 {
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
func (server *SystemgeServer) GetSyncSuccessResponsesReceived() uint64 {
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

func (server *SystemgeServer) CheckSyncFailureResponsesReceived() uint64 {
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
func (server *SystemgeServer) GetSyncFailureResponsesReceived() uint64 {
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

func (server *SystemgeServer) CheckNoSyncResponseReceived() uint64 {
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
func (server *SystemgeServer) GetNoSyncResponseReceived() uint64 {
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

func (server *SystemgeServer) CheckInvalidMessagesReceived() uint64 {
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
func (server *SystemgeServer) GetInvalidMessagesReceived() uint64 {
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

func (server *SystemgeServer) CheckSyncResponsesSent() uint64 {
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
func (server *SystemgeServer) GetSyncResponsesSent() uint64 {
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

func (server *SystemgeServer) CheckMessagesReceived() uint64 {
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
func (server *SystemgeServer) GetMessagesReceived() uint64 {
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

func (server *SystemgeServer) CheckRejectedMessages() uint64 {
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
func (server *SystemgeServer) GetRejectedMessages() uint64 {
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
