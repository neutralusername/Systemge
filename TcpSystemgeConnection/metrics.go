package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (connection *TcpSystemgeConnection) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcpSystemgeConnection_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     connection.CheckBytesSent(),
			"bytes_received": connection.CheckBytesReceived(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesSent", Metrics.New(
		map[string]uint64{
			"async_messages_sent": connection.CheckAsyncMessagesSent(),
			"sync_requests_sent":  connection.CheckSyncRequestsSent(),
			"sync_responses_sent": connection.CheckSyncResponsesSent(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			"sync_success_responses_received": connection.CheckSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": connection.CheckSyncFailureResponsesReceived(),
			"no_sync_response_received":       connection.CheckNoSyncResponseReceived(),
			"invalid_sync_responses_received": connection.CheckInvalidSyncResponsesReceived(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesReceived", Metrics.New(
		map[string]uint64{
			"invalid_messages_received": connection.CheckInvalidMessagesReceived(),
			"valid_messages_received":   connection.CheckValidMessagesReceived(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_rateLimiter", Metrics.New(
		map[string]uint64{
			"message_rate_limiter_exceeded": connection.CheckMessageRateLimiterExceeded(),
			"byte_rate_limiter_exceeded":    connection.CheckByteRateLimiterExceeded(),
		},
	))
	return metricsTypes
}

func (connection *TcpSystemgeConnection) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcpSystemgeConnection_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     connection.GetBytesSent(),
			"bytes_received": connection.GetBytesReceived(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesSent", Metrics.New(
		map[string]uint64{
			"async_messages_sent": connection.GetAsyncMessagesSent(),
			"sync_requests_sent":  connection.GetSyncRequestsSent(),
			"sync_responses_sent": connection.GetSyncResponsesSent(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_syncResponsesReceived", Metrics.New(
		map[string]uint64{
			"sync_success_responses_received": connection.GetSyncSuccessResponsesReceived(),
			"sync_failure_responses_received": connection.GetSyncFailureResponsesReceived(),
			"no_sync_response_received":       connection.GetNoSyncResponseReceived(),
			"invalid_sync_responses_received": connection.GetInvalidSyncResponsesReceived(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_messagesReceived", Metrics.New(
		map[string]uint64{
			"valid_messages_received":   connection.GetValidMessagesReceived(),
			"invalid_messages_received": connection.GetInvalidMessagesReceived(),
		},
	))
	metricsTypes.AddMetrics("tcpSystemgeConnection_rateLimiter", Metrics.New(
		map[string]uint64{
			"message_rate_limiter_exceeded": connection.GetMessageRateLimiterExceeded(),
			"byte_rate_limiter_exceeded":    connection.GetByteRateLimiterExceeded(),
		},
	))
	return metricsTypes
}

func (connection *TcpSystemgeConnection) CheckBytesSent() uint64 {
	return connection.bytesSent.Load()
}
func (connection *TcpSystemgeConnection) GetBytesSent() uint64 {
	return connection.bytesSent.Swap(0)
}

func (connection *TcpSystemgeConnection) CheckBytesReceived() uint64 {
	return connection.messageReceiver.CheckBytesReceived()
}
func (connection *TcpSystemgeConnection) GetBytesReceived() uint64 {
	return connection.messageReceiver.GetBytesReceived()
}

func (connection *TcpSystemgeConnection) CheckAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Load()
}
func (connection *TcpSystemgeConnection) GetAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Swap(0)
}

func (connection *TcpSystemgeConnection) CheckSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Load()
}
func (connection *TcpSystemgeConnection) GetSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Swap(0)
}

func (connection *TcpSystemgeConnection) CheckSyncResponsesSent() uint64 {
	return connection.syncResponsesSent.Load()
}
func (connection *TcpSystemgeConnection) GetSyncResponsesSent() uint64 {
	return connection.syncResponsesSent.Swap(0)
}

func (connection *TcpSystemgeConnection) CheckSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Load()
}
func (connection *TcpSystemgeConnection) GetSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Swap(0)
}

func (connection *TcpSystemgeConnection) CheckSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Load()
}
func (connection *TcpSystemgeConnection) GetSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Swap(0)
}

func (connection *TcpSystemgeConnection) CheckNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Load()
}
func (connection *TcpSystemgeConnection) GetNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Swap(0)
}

func (receiver *TcpSystemgeConnection) CheckInvalidSyncResponsesReceived() uint64 {
	return receiver.invalidSyncResponsesReceived.Load()
}
func (receiver *TcpSystemgeConnection) GetInvalidSyncResponsesReceived() uint64 {
	return receiver.invalidSyncResponsesReceived.Swap(0)
}

func (receiver *TcpSystemgeConnection) CheckMessagesReceived() uint64 {
	return receiver.messagesReceived.Load()
}
func (receiver *TcpSystemgeConnection) GetMessagesReceived() uint64 {
	return receiver.messagesReceived.Swap(0)
}

func (receiver *TcpSystemgeConnection) CheckInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Load()
}
func (receiver *TcpSystemgeConnection) GetInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Swap(0)
}

func (receiver *TcpSystemgeConnection) CheckRejectedMessages() uint64 {
	return receiver.rejectedMessages.Load()
}
func (receiver *TcpSystemgeConnection) GetRejectedMessages() uint64 {
	return receiver.rejectedMessages.Swap(0)
}
