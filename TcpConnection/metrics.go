package TcpConnection

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (connection *TcpConnection) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcpSystemgeConnection_byteTransmissions", Metrics.New(
		map[string]uint64{
			/* 	"bytes_sent":     connection.CheckBytesSent(),
			"bytes_received": connection.CheckBytesReceived(), */
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesSent", Metrics.New(
		map[string]uint64{
			/* "async_messages_sent": connection.CheckAsyncMessagesSent(),
			"sync_requests_sent":  connection.CheckSyncRequestsSent(),
			"sync_responses_sent": connection.CheckSyncResponsesSent(), */
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			/* 	"sync_success_responses_received": connection.CheckSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": connection.CheckSyncFailureResponsesReceived(),
			"no_sync_response_received":       connection.CheckNoSyncResponseReceived(), */
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesReceived", Metrics.New(
		map[string]uint64{
			/* "invalid_messages_received": connection.CheckInvalidMessagesReceived(),
			"messages_received":         connection.CheckMessagesReceived(),
			"rejected_messages":         connection.CheckRejectedMessages(), */
		},
	))
	return metricsTypes
}

func (connection *TcpConnection) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcpSystemgeConnection_byteTransmissions", Metrics.New(
		map[string]uint64{
			/* 	"bytes_sent":     connection.GetBytesSent(),
			"bytes_received": connection.GetBytesReceived(), */
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesSent", Metrics.New(
		map[string]uint64{
			/* "async_messages_sent": connection.GetAsyncMessagesSent(),
			"sync_requests_sent":  connection.GetSyncRequestsSent(),
			"sync_responses_sent": connection.GetSyncResponsesSent(), */
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			/* "sync_success_responses_received": connection.GetSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": connection.GetSyncFailureResponsesReceived(),
			"no_sync_response_received":       connection.GetNoSyncResponseReceived(), */
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesReceived", Metrics.New(
		map[string]uint64{
			/* 	"valid_messages_received":   connection.GetMessagesReceived(),
			"invalid_messages_received": connection.GetInvalidMessagesReceived(),
			"rejected_messages":         connection.GetRejectedMessages(), */
		},
	))
	return metricsTypes
}