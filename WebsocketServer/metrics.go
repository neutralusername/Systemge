package WebsocketServer

func (server *WebsocketServer) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytes_sent":         server.GetBytesSentCounter(),
		"bytes_received":     server.GetBytesReceivedCounter(),
		"incoming_messages":  uint64(server.GetIncomingMessageCounter()),
		"outgoing_messages":  uint64(server.GetOutgoingMessageCounter()),
		"active_connections": uint64(server.GetClientCount()),
	}
}

func (server *WebsocketServer) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"bytes_sent":         server.RetrieveBytesSentCounter(),
		"bytes_received":     server.RetrieveBytesReceivedCounter(),
		"incoming_messages":  uint64(server.RetrieveIncomingMessageCounter()),
		"outgoing_messages":  uint64(server.RetrieveOutgoingMessageCounter()),
		"active_connections": uint64(server.GetClientCount()),
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
