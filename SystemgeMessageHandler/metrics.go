package SystemgeMessageHandler

func (messageHandler *SystemgeMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}

func (messageHandler *SystemgeMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}

func (messageHandler *SystemgeMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}

func (messageHandler *SystemgeMessageHandler) RetrieveAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *SystemgeMessageHandler) RetrieveSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *SystemgeMessageHandler) RetrieveUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
