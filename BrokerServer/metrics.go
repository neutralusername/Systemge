package BrokerServer

func (server *Server) GetAsyncMessagesReceived() uint64 {
	return server.asyncMessagesReceived.Load()
}
func (server *Server) RetrieveAsyncMessagesReceived() uint64 {
	return server.asyncMessagesReceived.Swap(0)
}

func (server *Server) GetAsyncMessagesPropagated() uint64 {
	return server.asyncMessagesPropagated.Load()
}
func (server *Server) RetrieveAsyncMessagesPropagated() uint64 {
	return server.asyncMessagesPropagated.Swap(0)
}

func (server *Server) GetSyncRequestsReceived() uint64 {
	return server.syncRequestsReceived.Load()
}
func (server *Server) RetrieveSyncRequestsReceived() uint64 {
	return server.syncRequestsReceived.Swap(0)
}

func (server *Server) GetSyncRequestsPropagated() uint64 {
	return server.syncRequestsPropagated.Load()
}
func (server *Server) RetrieveSyncRequestsPropagated() uint64 {
	return server.syncRequestsPropagated.Swap(0)
}
