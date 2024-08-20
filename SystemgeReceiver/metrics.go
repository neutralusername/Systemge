package SystemgeReceiver

func (receiver *SystemgeReceiver) GetAsyncMessagesReceived() uint64 {
	return receiver.asyncMessagesReceived.Load()
}

func (receiver *SystemgeReceiver) GetSyncRequestsReceived() uint64 {
	return receiver.syncRequestsReceived.Load()
}

func (receiver *SystemgeReceiver) GetInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Load()
}

func (receiver *SystemgeReceiver) GetSyncResponsesReceived() uint64 {
	return receiver.syncResponsesReceived.Load()
}

func (receiver *SystemgeReceiver) GetMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Load()
}

func (receiver *SystemgeReceiver) GetByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Load()
}

func (receiver *SystemgeReceiver) RetrieveAsyncMessagesReceived() uint64 {
	return receiver.asyncMessagesReceived.Swap(0)
}

func (receiver *SystemgeReceiver) RetrieveSyncRequestsReceived() uint64 {
	return receiver.syncRequestsReceived.Swap(0)
}

func (receiver *SystemgeReceiver) RetrieveInvalidMessagesReceived() uint64 {
	return receiver.invalidMessagesReceived.Swap(0)
}

func (receiver *SystemgeReceiver) RetrieveSyncResponsesReceived() uint64 {
	return receiver.syncResponsesReceived.Swap(0)
}

func (receiver *SystemgeReceiver) RetrieveMessageRateLimiterExceeded() uint64 {
	return receiver.messageRateLimiterExceeded.Swap(0)
}

func (receiver *SystemgeReceiver) RetrieveByteRateLimiterExceeded() uint64 {
	return receiver.byteRateLimiterExceeded.Swap(0)
}
