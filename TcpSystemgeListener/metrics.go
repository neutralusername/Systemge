package TcpSystemgeListener

func (listener *TcpListener) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"connection_attempts":  listener.GetConnectionAttempts(),
		"failed_connections":   listener.GetFailedConnections(),
		"rejected_connections": listener.GetRejectedConnections(),
		"accepted_connections": listener.GetAcceptedConnections(),
	}
}
func (listener *TcpListener) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"connection_attempts":  listener.RetrieveConnectionAttempts(),
		"failed_connections":   listener.RetrieveFailedConnections(),
		"rejected_connections": listener.RetrieveRejectedConnections(),
		"accepted_connections": listener.RetrieveAcceptedConnections(),
	}
}

func (listener *TcpListener) GetConnectionAttempts() uint64 {
	return listener.connectionAttempts.Load()
}
func (listener *TcpListener) RetrieveConnectionAttempts() uint64 {
	return listener.connectionAttempts.Swap(0)
}

func (listener *TcpListener) GetFailedConnections() uint64 {
	return listener.failedConnections.Load()
}
func (listener *TcpListener) RetrieveFailedConnections() uint64 {
	return listener.failedConnections.Swap(0)
}

func (listener *TcpListener) GetRejectedConnections() uint64 {
	return listener.rejectedConnections.Load()
}
func (listener *TcpListener) RetrieveRejectedConnections() uint64 {
	return listener.rejectedConnections.Swap(0)
}

func (listener *TcpListener) GetAcceptedConnections() uint64 {
	return listener.acceptedConnections.Load()
}
func (listener *TcpListener) RetrieveAcceptedConnections() uint64 {
	return listener.acceptedConnections.Swap(0)
}
