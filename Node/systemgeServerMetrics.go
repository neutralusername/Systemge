package Node

func (node *Node) RetrieveSystemgeServerMessageRateLimiterExceeded() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.messageRateLimiterExceeded.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerByteRateLimiterExceeded() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.byteRateLimiterExceeded.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerConnectionAttempts() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.connectionAttempts.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerConnectionAttemptsSuccessful() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.connectionAttemptsSuccessful.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerConnectionAttemptsFailed() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.connectionAttemptsFailed.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerConnectionAttemptBytesSent() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.connectionAttemptBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerConnectionAttemptBytesReceived() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.connectionAttemptBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerInvalidMessagesReceived() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.invalidMessagesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerSyncSuccessResponsesSent() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.syncSuccessResponsesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerSyncFailureResponsesSent() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.syncFailureResponsesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerSyncResponseBytesSent() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.syncResponseBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerAsyncMessagesReceived() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.asyncMessagesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerAsyncMessageBytesReceived() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.asyncMessageBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerSyncRequestsReceived() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.syncRequestsReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerSyncRequestBytesReceived() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.syncRequestBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerTopicAddSent() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.topicAddSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerTopicRemoveSent() uint32 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.topicRemoveSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerBytesReceived() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.bytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeServerBytesSent() uint64 {
	if systemgeServer := node.systemgeServer; systemgeServer != nil {
		return systemgeServer.bytesSent.Swap(0)
	}
	return 0
}
