package WebsocketServer

func (server *WebsocketServer) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytesSent":         server.GetBytesSentCounter(),
		"bytesReceived":     server.GetBytesReceivedCounter(),
		"incomingMessages":  uint64(server.GetIncomingMessageCounter()),
		"outgoingMessages":  uint64(server.GetOutgoingMessageCounter()),
		"activeConnections": uint64(server.GetClientCount()),
	}
}

func (server *WebsocketServer) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytesSent":         server.RetrieveBytesSentCounter(),
		"bytesReceived":     server.RetrieveBytesReceivedCounter(),
		"incomingMessages":  uint64(server.RetrieveIncomingMessageCounter()),
		"outgoingMessages":  uint64(server.RetrieveOutgoingMessageCounter()),
		"activeConnections": uint64(server.GetClientCount()),
	}
}

func (server *WebsocketServer) RetrieveBytesSentCounter() uint64 {
	return server.bytesSentCounter.Swap(0)
}
func (server *WebsocketServer) GetBytesSentCounter() uint64 {
	return server.bytesSentCounter.Load()
}

func (server *WebsocketServer) RetrieveBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Swap(0)
}
func (server *WebsocketServer) GetBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Load()
}

func (server *WebsocketServer) RetrieveIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Swap(0)
}
func (server *WebsocketServer) GetIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Load()
}

func (server *WebsocketServer) RetrieveOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Swap(0)
}
func (server *WebsocketServer) GetOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Load()
}
