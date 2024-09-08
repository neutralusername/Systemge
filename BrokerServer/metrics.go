package BrokerServer

func (server *Server) GetMetrics() map[string]uint64 {
	metrics := map[string]uint64{}
	metrics["async_messages_received"] = server.GetAsyncMessagesReceived()
	metrics["async_messages_propagated"] = server.GetAsyncMessagesPropagated()
	metrics["sync_requests_received"] = server.GetSyncRequestsReceived()
	metrics["sync_requests_propagated"] = server.GetSyncRequestsPropagated()
	metrics["connection_count"] = uint64(len(server.connectionAsyncSubscriptions))
	metrics["async_topic_count"] = uint64(len(server.asyncTopicSubscriptions))
	metrics["sync_topic_count"] = uint64(len(server.syncTopicSubscriptions))
	systemgeServerMetrics := server.systemgeServer.GetMetrics()
	for key, value := range systemgeServerMetrics {
		metrics[key] = value
	}
	messageHandlerMetrics := server.messageHandler.GetMetrics()
	for key, value := range messageHandlerMetrics {
		metrics[key] = value
	}
	return metrics
}
func (server *Server) RetrieveMetrics() map[string]uint64 {
	metrics := map[string]uint64{}
	metrics["async_messages_received"] = server.RetrieveAsyncMessagesPropagated()
	metrics["async_messages_propagated"] = server.RetrieveAsyncMessagesPropagated()
	metrics["sync_requests_received"] = server.RetrieveSyncRequestsReceived()
	metrics["sync_requests_propagated"] = server.RetrieveSyncRequestsPropagated()
	metrics["connection_count"] = uint64(len(server.connectionAsyncSubscriptions))
	metrics["async_topic_count"] = uint64(len(server.asyncTopicSubscriptions))
	metrics["sync_topic_count"] = uint64(len(server.syncTopicSubscriptions))
	systemgeServerMetrics := server.systemgeServer.RetrieveMetrics()
	for key, value := range systemgeServerMetrics {
		metrics[key] = value
	}
	messageHandlerMetrics := server.messageHandler.RetrieveMetrics()
	for key, value := range messageHandlerMetrics {
		metrics[key] = value
	}
	return metrics
}

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
