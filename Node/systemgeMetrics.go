package Node

func (node *Node) RetrieveOutgoingConnectionAttempts() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttempts.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingConnectionAttemptsSuccessful() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptsSuccessful.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingConnectionAttemptsFailed() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptsFailed.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingConnectionAttemptBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingConnectionAttemptBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveInvalidMessagesFromOutgoingConnections() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.invalidMessagesFromOutgoingConnections.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingSyncResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncResponses.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingSyncSuccessResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncSuccessResponses.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingSyncFailureResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncFailureResponses.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingSyncResponseBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncResponseBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingAsyncMessages() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingAsyncMessages.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingAsyncMessageBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingAsyncMessageBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingSyncRequests() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncRequests.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingSyncRequestBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncRequestBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingConnectionAttempts() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttempts.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingConnectionAttemptsSuccessful() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptsSuccessful.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingConnectionAttemptsFailed() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptsFailed.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingConnectionAttemptBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingConnectionAttemptBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveInvalidMessagesFromIncomingConnections() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.invalidMessagesFromIncomingConnections.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingSyncRequests() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncRequests.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingSyncRequestBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncRequestBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingAsyncMessages() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingAsyncMessages.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveIncomingAsyncMessageBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingAsyncMessageBytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingSyncResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncResponses.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingSyncSuccessResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncSuccessResponses.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingSyncFailureResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncFailureResponses.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveOutgoingSyncResponseBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncResponseBytesSent.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeBytesReceivedCounter() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.bytesReceived.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveSystemgeBytesSentCounter() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.bytesSent.Swap(0)
	}
	return 0
}

func (node *Node) GetOutgoingConnectionAttempts() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttempts.Load()
	}
	return 0
}

func (node *Node) GetOutgoingConnectionAttemptsSuccessful() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptsSuccessful.Load()
	}
	return 0
}

func (node *Node) GetOutgoingConnectionAttemptsFailed() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptsFailed.Load()
	}
	return 0
}

func (node *Node) GetOutgoingConnectionAttemptBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptBytesSent.Load()
	}
	return 0
}

func (node *Node) GetOutgoingConnectionAttemptBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingConnectionAttemptBytesReceived.Load()
	}
	return 0
}

func (node *Node) GetInvalidMessagesFromOutgoingConnections() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.invalidMessagesFromOutgoingConnections.Load()
	}
	return 0
}

func (node *Node) GetIncomingSyncResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncResponses.Load()
	}
	return 0
}

func (node *Node) GetIncomingSyncSuccessResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncSuccessResponses.Load()
	}
	return 0
}

func (node *Node) GetIncomingSyncFailureResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncFailureResponses.Load()
	}
	return 0
}

func (node *Node) GetIncomingSyncResponseBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncResponseBytesReceived.Load()
	}
	return 0
}

func (node *Node) GetOutgoingAsyncMessages() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingAsyncMessages.Load()
	}
	return 0
}

func (node *Node) GetOutgoingAsyncMessageBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingAsyncMessageBytesSent.Load()
	}
	return 0
}

func (node *Node) GetOutgoingSyncRequests() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncRequests.Load()
	}
	return 0
}

func (node *Node) GetOutgoingSyncRequestBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncRequestBytesSent.Load()
	}
	return 0
}

func (node *Node) GetIncomingConnectionAttempts() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttempts.Load()
	}
	return 0
}

func (node *Node) GetIncomingConnectionAttemptsSuccessful() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptsSuccessful.Load()
	}
	return 0
}

func (node *Node) GetIncomingConnectionAttemptsFailed() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptsFailed.Load()
	}
	return 0
}

func (node *Node) GetIncomingConnectionAttemptBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptBytesSent.Load()
	}
	return 0
}

func (node *Node) GetIncomingConnectionAttemptBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingConnectionAttemptBytesReceived.Load()
	}
	return 0
}

func (node *Node) GetInvalidMessagesFromIncomingConnections() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.invalidMessagesFromIncomingConnections.Load()
	}
	return 0
}

func (node *Node) GetIncomingSyncRequests() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncRequests.Load()
	}
	return 0
}

func (node *Node) GetIncomingSyncRequestBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingSyncRequestBytesReceived.Load()
	}
	return 0
}

func (node *Node) GetIncomingAsyncMessages() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingAsyncMessages.Load()
	}
	return 0
}

func (node *Node) GetIncomingAsyncMessageBytesReceived() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.incomingAsyncMessageBytesReceived.Load()
	}
	return 0
}

func (node *Node) GetOutgoingSyncResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncResponses.Load()
	}
	return 0
}

func (node *Node) GetOutgoingSyncSuccessResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncSuccessResponses.Load()
	}
	return 0
}

func (node *Node) GetOutgoingSyncFailureResponses() uint32 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncFailureResponses.Load()
	}
	return 0
}

func (node *Node) GetOutgoingSyncResponseBytesSent() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.outgoingSyncResponseBytesSent.Load()
	}
	return 0
}

func (node *Node) GetSystemgeBytesReceivedCounter() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.bytesReceived.Load()
	}
	return 0
}

func (node *Node) GetSystemgeBytesSentCounter() uint64 {
	if systemge := node.systemge; systemge != nil {
		return systemge.bytesSent.Load()
	}
	return 0
}
