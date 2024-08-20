package SystemgeConnection

func (connection *SystemgeConnection) GetBytesSent() uint64 {
	return connection.bytesSent.Load()
}

func (connection *SystemgeConnection) GetBytesReceived() uint64 {
	return connection.bytesReceived.Load()
}

func (connection *SystemgeConnection) GetAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Load()
}

func (connection *SystemgeConnection) GetSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Load()
}

func (connection *SystemgeConnection) GetSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Load()
}

func (connection *SystemgeConnection) GetSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Load()
}

func (connection *SystemgeConnection) GetNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Load()
}

func (connection *SystemgeConnection) RetrieveBytesSent() uint64 {
	return connection.bytesSent.Swap(0)
}

func (connection *SystemgeConnection) RetrieveBytesReceived() uint64 {
	return connection.bytesReceived.Swap(0)
}

func (connection *SystemgeConnection) RetrieveAsyncMessagesSent() uint64 {
	return connection.asyncMessagesSent.Swap(0)
}

func (connection *SystemgeConnection) RetrieveSyncRequestsSent() uint64 {
	return connection.syncRequestsSent.Swap(0)
}

func (connection *SystemgeConnection) RetrieveSyncSuccessResponsesReceived() uint64 {
	return connection.syncSuccessResponsesReceived.Swap(0)
}

func (connection *SystemgeConnection) RetrieveSyncFailureResponsesReceived() uint64 {
	return connection.syncFailureResponsesReceived.Swap(0)
}

func (connection *SystemgeConnection) RetrieveNoSyncResponseReceived() uint64 {
	return connection.noSyncResponseReceived.Swap(0)
}

func (receiver *SystemgeConnection) GetAsyncMessagesReceived() uint64 {
	return receiver.asyncMessagesReceived.Load()
}

func (receiver *SystemgeConnection) GetSyncRequestsReceived() uint64 {
	return receiver.syncRequestsReceived.Load()
}

func (receiver *SystemgeConnection) GetInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Load()
}

func (receiver *SystemgeConnection) GetMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Load()
}

func (receiver *SystemgeConnection) GetByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Load()
}

func (receiver *SystemgeConnection) RetrieveAsyncMessagesReceived() uint64 {
	return receiver.asyncMessagesReceived.Swap(0)
}

func (receiver *SystemgeConnection) RetrieveSyncRequestsReceived() uint64 {
	return receiver.syncRequestsReceived.Swap(0)
}

func (receiver *SystemgeConnection) RetrieveInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Swap(0)
}

func (receiver *SystemgeConnection) RetrieveMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Swap(0)
}

func (receiver *SystemgeConnection) RetrieveByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Swap(0)
}
