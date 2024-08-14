package Node

func (client *SystemgeClient) RetrieveSystemgeClientMessageRateLimiterExceeded() uint32 {
	return client.messageRateLimiterExceeded.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientByteRateLimiterExceeded() uint32 {
	return client.byteRateLimiterExceeded.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientConnectionAttempts() uint32 {
	return client.connectionAttempts.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientConnectionAttemptsSuccessful() uint32 {
	return client.connectionAttemptsSuccessful.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientConnectionAttemptsFailed() uint32 {
	return client.connectionAttemptsFailed.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientConnectionAttemptBytesSent() uint64 {
	return client.connectionAttemptBytesSent.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientConnectionAttemptBytesReceived() uint64 {
	return client.connectionAttemptBytesReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientInvalidMessagesReceived() uint32 {
	return client.invalidMessagesReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientSyncSuccessResponsesReceived() uint32 {
	return client.syncSuccessResponsesReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientSyncFailureResponsesReceived() uint32 {
	return client.syncFailureResponsesReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientSyncResponseBytesReceived() uint64 {
	return client.syncResponseBytesReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientAsyncMessagesSent() uint32 {
	return client.asyncMessagesSent.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientAsyncMessageBytesSent() uint64 {
	return client.asyncMessageBytesSent.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientSyncRequestsSent() uint32 {
	return client.syncRequestsSent.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientSyncRequestBytesSent() uint64 {
	return client.syncRequestBytesSent.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientTopicAddReceived() uint32 {
	return client.topicAddReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientTopicRemoveReceived() uint32 {
	return client.topicRemoveReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientBytesReceived() uint64 {
	return client.bytesReceived.Swap(0)
}

func (client *SystemgeClient) RetrieveSystemgeClientBytesSent() uint64 {
	return client.bytesSent.Swap(0)
}
