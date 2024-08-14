package WebsocketServer

func (server *Server) RetrieveWebsocketBytesSentCounter() uint64 {
	return server.bytesSentCounter.Swap(0)
}
func (server *Server) GetWebsocketBytesSentCounter() uint64 {
	return server.bytesSentCounter.Load()
}

func (server *Server) RetrieveWebsocketBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Swap(0)
}
func (server *Server) GetWebsocketBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Load()
}

func (server *Server) RetrieveWebsocketIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Swap(0)
}
func (server *Server) GetWebsocketIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Load()
}

func (server *Server) RetrieveWebsocketOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Swap(0)
}
func (server *Server) GetWebsocketOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Load()
}
