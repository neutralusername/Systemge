package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

func (connection *TcpConnection) CheckMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"tcp_systemge_connection": {
			KeyValuePairs: map[string]uint64{
				"bytes_sent":                      connection.CheckBytesSent(),
				"bytes_received":                  connection.CheckBytesReceived(),
				"async_messages_sent":             connection.CheckAsyncMessagesSent(),
				"sync_requests_sent":              connection.CheckSyncRequestsSent(),
				"sync_success_responses_received": connection.CheckSyncSuccessResponsesReceived(),
				"sync_failure_responses_received": connection.CheckSyncFailureResponsesReceived(),
				"no_sync_response_received":       connection.CheckNoSyncResponseReceived(),
				"invalid_messages_received":       connection.CheckInvalidMessagesReceived(),
				"invalid_sync_responses_received": connection.CheckInvalidSyncResponsesReceived(),
				"valid_messages_received":         connection.CheckValidMessagesReceived(),
				"message_rate_limiter_exceeded":   connection.CheckMessageRateLimiterExceeded(),
				"byte_rate_limiter_exceeded":      connection.CheckByteRateLimiterExceeded(),
			},
			Time: time.Now(),
		},
	}
}

func (connection *TcpConnection) GetMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"tcp_systemge_connection": {
			KeyValuePairs: map[string]uint64{
				"bytes_sent":                      connection.GetBytesSent(),
				"bytes_received":                  connection.GetBytesReceived(),
				"async_messages_sent":             connection.GetAsyncMessagesSent(),
				"sync_requests_sent":              connection.GetSyncRequestsSent(),
				"sync_success_responses_received": connection.GetSyncSuccessResponsesReceived(),
				"sync_failure_responses_received": connection.GetSyncFailureResponsesReceived(),
				"no_sync_response_received":       connection.GetNoSyncResponseReceived(),
				"invalid_messages_received":       connection.GetInvalidMessagesReceived(),
				"invalid_sync_responses_received": connection.GetInvalidSyncResponsesReceived(),
				"valid_messages_received":         connection.GetValidMessagesReceived(),
				"message_rate_limiter_exceeded":   connection.GetMessageRateLimiterExceeded(),
				"byte_rate_limiter_exceeded":      connection.GetByteRateLimiterExceeded(),
			},
			Time: time.Now(),
		},
	}
}

func (connection *TcpConnection) CheckBytesSent() uint64 {
	return connection.bytesSent.Load()
}
func (connection *TcpConnection) GetBytesSent() uint64 {
	return connection.bytesSent.Swap(0)
}

func (connection *TcpConnection) CheckBytesReceived() uint64 {
	return connection.bytesReceived.Load()
}
func (connection *TcpConnection) GetBytesReceived() uint64 {
	return connection.bytesReceived.Swap(0)
}

func (connection *TcpConnection) CheckAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Load()
}
func (connection *TcpConnection) GetAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Swap(0)
}

func (connection *TcpConnection) CheckSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Load()
}
func (connection *TcpConnection) GetSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Swap(0)
}

func (connection *TcpConnection) CheckSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Load()
}
func (connection *TcpConnection) GetSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Swap(0)
}

func (connection *TcpConnection) CheckSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Load()
}
func (connection *TcpConnection) GetSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Swap(0)
}

func (connection *TcpConnection) CheckNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Load()
}
func (connection *TcpConnection) GetNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Swap(0)
}

func (receiver *TcpConnection) CheckInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Load()
}
func (receiver *TcpConnection) GetInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Swap(0)
}

func (receiver *TcpConnection) CheckInvalidSyncResponsesReceived() uint64 {
	return receiver.invalidSyncResponsesReceived.Load()
}
func (receiver *TcpConnection) GetInvalidSyncResponsesReceived() uint64 {
	return receiver.invalidSyncResponsesReceived.Swap(0)
}

func (receiver *TcpConnection) CheckValidMessagesReceived() uint64 {
	return receiver.validMessagesReceived.Load()
}
func (receiver *TcpConnection) GetValidMessagesReceived() uint64 {
	return receiver.validMessagesReceived.Swap(0)
}

func (receiver *TcpConnection) CheckMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Load()
}
func (receiver *TcpConnection) GetMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Swap(0)
}

func (receiver *TcpConnection) CheckByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Load()
}
func (receiver *TcpConnection) GetByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Swap(0)
}
