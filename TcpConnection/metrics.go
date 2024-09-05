package TcpConnection

func (connection *TcpConnection) GetMetrics() map[string]uint64 {
	return map[string]uint64{
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
	}
}

func (connection *TcpConnection) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytes_sent":                      connection.RetrieveBytesSent(),
		"bytes_received":                  connection.RetrieveBytesReceived(),
		"async_messages_sent":             connection.RetrieveAsyncMessagesSent(),
		"sync_requests_sent":              connection.RetrieveSyncRequestsSent(),
		"sync_success_responses_received": connection.RetrieveSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": connection.RetrieveSyncFailureResponsesReceived(),
		"no_sync_response_received":       connection.RetrieveNoSyncResponseReceived(),
		"invalid_messages_received":       connection.RetrieveInvalidMessagesReceived(),
		"invalid_sync_responses_received": connection.RetrieveInvalidSyncResponsesReceived(),
		"valid_messages_received":         connection.RetrieveValidMessagesReceived(),
		"message_rate_limiter_exceeded":   connection.RetrieveMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      connection.RetrieveByteRateLimiterExceeded(),
	}
}

func (connection *TcpConnection) GetBytesSent() uint64 {
	return connection.bytesSent.Load()
}
func (connection *TcpConnection) RetrieveBytesSent() uint64 {
	return connection.bytesSent.Swap(0)
}

func (connection *TcpConnection) GetBytesReceived() uint64 {
	return connection.bytesReceived.Load()
}
func (connection *TcpConnection) RetrieveBytesReceived() uint64 {
	return connection.bytesReceived.Swap(0)
}

func (connection *TcpConnection) GetAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Load()
}
func (connection *TcpConnection) RetrieveAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Swap(0)
}

func (connection *TcpConnection) GetSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Load()
}
func (connection *TcpConnection) RetrieveSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Swap(0)
}

func (connection *TcpConnection) GetSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Load()
}
func (connection *TcpConnection) RetrieveSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Swap(0)
}

func (connection *TcpConnection) GetSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Load()
}
func (connection *TcpConnection) RetrieveSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Swap(0)
}

func (connection *TcpConnection) GetNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Load()
}
func (connection *TcpConnection) RetrieveNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Swap(0)
}

func (receiver *TcpConnection) GetInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Load()
}
func (receiver *TcpConnection) RetrieveInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Swap(0)
}

func (receiver *TcpConnection) GetInvalidSyncResponsesReceived() uint64 {
	return receiver.invalidSyncResponsesReceived.Load()
}
func (receiver *TcpConnection) RetrieveInvalidSyncResponsesReceived() uint64 {
	return receiver.invalidSyncResponsesReceived.Swap(0)
}

func (receiver *TcpConnection) GetValidMessagesReceived() uint64 {
	return receiver.validMessagesReceived.Load()
}
func (receiver *TcpConnection) RetrieveValidMessagesReceived() uint64 {
	return receiver.validMessagesReceived.Swap(0)
}

func (receiver *TcpConnection) GetMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Load()
}
func (receiver *TcpConnection) RetrieveMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Swap(0)
}

func (receiver *TcpConnection) GetByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Load()
}
func (receiver *TcpConnection) RetrieveByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Swap(0)
}
