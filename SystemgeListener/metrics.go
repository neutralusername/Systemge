package SystemgeListener

func (listener *SystemgeListener) GetConnectionAttempts() uint64 {
	return listener.connectionAttempts.Load()
}
func (listener *SystemgeListener) RetrieveConnectionAttempts() uint64 {
	return listener.connectionAttempts.Swap(0)
}

func (listener *SystemgeListener) GetFailedConnections() uint64 {
	return listener.failedConnections.Load()
}
func (listener *SystemgeListener) RetrieveFailedConnections() uint64 {
	return listener.failedConnections.Swap(0)
}

func (listener *SystemgeListener) GetRejectedConnections() uint64 {
	return listener.rejectedConnections.Load()
}
func (listener *SystemgeListener) RetrieveRejectedConnections() uint64 {
	return listener.rejectedConnections.Swap(0)
}

func (listener *SystemgeListener) GetAcceptedConnections() uint64 {
	return listener.acceptedConnections.Load()
}
func (listener *SystemgeListener) RetrieveAcceptedConnections() uint64 {
	return listener.acceptedConnections.Swap(0)
}
