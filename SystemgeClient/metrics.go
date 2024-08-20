package SystemgeClient

func (client *SystemgeClient) GetConnectionAttemptsFailed() uint32 {
	return client.connectionAttemptsFailed.Load()
}

func (client *SystemgeClient) GetConnectionAttemptsSuccess() uint32 {
	return client.connectionAttemptsSuccess.Load()
}

func (client *SystemgeClient) RetrieveConnectionAttemptsFailed() uint32 {
	return client.connectionAttemptsFailed.Swap(0)
}

func (client *SystemgeClient) RetrieveConnectionAttemptsSuccess() uint32 {
	return client.connectionAttemptsSuccess.Swap(0)
}
