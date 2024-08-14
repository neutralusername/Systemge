package SystemgeServer

func (server *SystemgeServer) RetrieveSystemgeServerMessageRateLimiterExceeded() uint32 {
	return server.messageRateLimiterExceeded.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerByteRateLimiterExceeded() uint32 {
	return server.byteRateLimiterExceeded.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerConnectionAttempts() uint32 {
	return server.connectionAttempts.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerConnectionAttemptsSuccessful() uint32 {
	return server.connectionAttemptsSuccessful.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerConnectionAttemptsFailed() uint32 {
	return server.connectionAttemptsFailed.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerConnectionAttemptBytesSent() uint64 {
	return server.connectionAttemptBytesSent.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerConnectionAttemptBytesReceived() uint64 {
	return server.connectionAttemptBytesReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerInvalidMessagesReceived() uint32 {
	return server.invalidMessagesReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerSyncSuccessResponsesSent() uint32 {
	return server.syncSuccessResponsesSent.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerSyncFailureResponsesSent() uint32 {
	return server.syncFailureResponsesSent.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerSyncResponseBytesSent() uint64 {
	return server.syncResponseBytesSent.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerAsyncMessagesReceived() uint32 {
	return server.asyncMessagesReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerAsyncMessageBytesReceived() uint64 {
	return server.asyncMessageBytesReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerSyncRequestsReceived() uint32 {
	return server.syncRequestsReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerSyncRequestBytesReceived() uint64 {
	return server.syncRequestBytesReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerTopicAddSent() uint32 {
	return server.topicAddSent.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerTopicRemoveSent() uint32 {
	return server.topicRemoveSent.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerBytesReceived() uint64 {
	return server.bytesReceived.Swap(0)
}

func (server *SystemgeServer) RetrieveSystemgeServerBytesSent() uint64 {
	return server.bytesSent.Swap(0)
}
