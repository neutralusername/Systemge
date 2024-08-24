package WebsocketServer

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
