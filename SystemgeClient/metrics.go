package Node

func (node *Node) RetrieveSystemgeClientMessageRateLimiterExceeded() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.messageRateLimiterExceeded.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientByteRateLimiterExceeded() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.byteRateLimiterExceeded.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientConnectionAttempts() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.connectionAttempts.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientConnectionAttemptsSuccessful() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.connectionAttemptsSuccessful.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientConnectionAttemptsFailed() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.connectionAttemptsFailed.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientConnectionAttemptBytesSent() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.connectionAttemptBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientConnectionAttemptBytesReceived() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.connectionAttemptBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientInvalidMessagesReceived() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.invalidMessagesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientSyncSuccessResponsesReceived() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.syncSuccessResponsesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientSyncFailureResponsesReceived() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.syncFailureResponsesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientSyncResponseBytesReceived() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.syncResponseBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientAsyncMessagesSent() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.asyncMessagesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientAsyncMessageBytesSent() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.asyncMessageBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientSyncRequestsSent() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.syncRequestsSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientSyncRequestBytesSent() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.syncRequestBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientTopicAddReceived() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.topicAddReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientTopicRemoveReceived() uint32 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.topicRemoveReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientBytesReceived() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.bytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeClientBytesSent() uint64 {
	if systemgeClient := node.systemgeClient; systemgeClient != nil {
		return systemgeClient.bytesSent.Swap(0)
	}
	return 0
}
