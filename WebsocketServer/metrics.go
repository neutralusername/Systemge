package WebsocketServer

func (server *WebsocketServer) RetrieveWebsocketBytesSentCounter() uint64 {
	return server.bytesSentCounter.Swap(0)
}
func (server *WebsocketServer) GetWebsocketBytesSentCounter() uint64 {
	return server.bytesSentCounter.Load()
}

func (server *WebsocketServer) RetrieveWebsocketBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Swap(0)
}
func (server *WebsocketServer) GetWebsocketBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Load()
}

func (server *WebsocketServer) RetrieveWebsocketIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Swap(0)
}
func (server *WebsocketServer) GetWebsocketIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Load()
}

func (server *WebsocketServer) RetrieveWebsocketOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Swap(0)
}
func (server *WebsocketServer) GetWebsocketOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Load()
}
