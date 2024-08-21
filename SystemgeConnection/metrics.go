package SystemgeConnection

func (connection *SystemgeConnection) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytes_sent":                      connection.GetBytesSent(),
		"bytes_received":                  connection.GetBytesReceived(),
		"async_messages_sent":             connection.GetAsyncMessagesSent(),
		"sync_requests_sent":              connection.GetSyncRequestsSent(),
		"sync_success_responses_received": connection.GetSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": connection.GetSyncFailureResponsesReceived(),
		"no_sync_response_received":       connection.GetNoSyncResponseReceived(),
		"async_messages_received":         connection.GetAsyncMessagesReceived(),
		"sync_requests_received":          connection.GetSyncRequestsReceived(),
		"invalid_messages_received":       connection.GetInvalidMessagesReceived(),
		"message_rate_limiter_exceeded":   connection.GetMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      connection.GetByteRateLimiterExceeded(),
	}
}

func (connection *SystemgeConnection) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytes_sent":                      connection.RetrieveBytesSent(),
		"bytes_received":                  connection.RetrieveBytesReceived(),
		"async_messages_sent":             connection.RetrieveAsyncMessagesSent(),
		"sync_requests_sent":              connection.RetrieveSyncRequestsSent(),
		"sync_success_responses_received": connection.RetrieveSyncSuccessResponsesReceived(),
		"sync_failure_responses_received": connection.RetrieveSyncFailureResponsesReceived(),
		"no_sync_response_received":       connection.RetrieveNoSyncResponseReceived(),
		"async_messages_received":         connection.RetrieveAsyncMessagesReceived(),
		"sync_requests_received":          connection.RetrieveSyncRequestsReceived(),
		"invalid_messages_received":       connection.RetrieveInvalidMessagesReceived(),
		"message_rate_limiter_exceeded":   connection.RetrieveMessageRateLimiterExceeded(),
		"byte_rate_limiter_exceeded":      connection.RetrieveByteRateLimiterExceeded(),
	}
}

func (connection *SystemgeConnection) GetBytesSent() uint64 {
	return connection.bytesSent.Load()
}
func (connection *SystemgeConnection) RetrieveBytesSent() uint64 {
	return connection.bytesSent.Swap(0)
}

func (connection *SystemgeConnection) GetBytesReceived() uint64 {
	return connection.bytesReceived.Load()
}
func (connection *SystemgeConnection) RetrieveBytesReceived() uint64 {
	return connection.bytesReceived.Swap(0)
}

func (connection *SystemgeConnection) GetAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Load()
}
func (connection *SystemgeConnection) RetrieveAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Swap(0)
}

func (connection *SystemgeConnection) GetSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Load()
}
func (connection *SystemgeConnection) RetrieveSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Swap(0)
}

func (connection *SystemgeConnection) GetSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Load()
}
func (connection *SystemgeConnection) RetrieveSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Swap(0)
}

func (connection *SystemgeConnection) GetSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Load()
}
func (connection *SystemgeConnection) RetrieveSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Swap(0)
}

func (connection *SystemgeConnection) GetNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Load()
}
func (connection *SystemgeConnection) RetrieveNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Swap(0)
}

func (receiver *SystemgeConnection) GetAsyncMessagesReceived() uint64 {
	return receiver.asyncMessagesReceived.Load()
}
func (receiver *SystemgeConnection) RetrieveAsyncMessagesReceived() uint64 {
	return receiver.asyncMessagesReceived.Swap(0)
}

func (receiver *SystemgeConnection) GetSyncRequestsReceived() uint64 {
	return receiver.syncRequestsReceived.Load()
}
func (receiver *SystemgeConnection) RetrieveSyncRequestsReceived() uint64 {
	return receiver.syncRequestsReceived.Swap(0)
}

func (receiver *SystemgeConnection) GetInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Load()
}
func (receiver *SystemgeConnection) RetrieveInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Swap(0)
}

func (receiver *SystemgeConnection) GetMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Load()
}
func (receiver *SystemgeConnection) RetrieveMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Swap(0)
}

func (receiver *SystemgeConnection) GetByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Load()
}
func (receiver *SystemgeConnection) RetrieveByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Swap(0)
}
