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
