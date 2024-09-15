package BrokerServer

func (server *Server) CheckMetrics() map[string]map[string]uint64 {
	metrics := map[string]map[string]uint64{}
	metrics["broker_server"] = map[string]uint64{}
	server.mutex.Lock()
	metrics["broker_server"]["async_messages_received"] = server.CheckAsyncMessagesReceived()
	metrics["broker_server"]["async_messages_propagated"] = server.CheckAsyncMessagesPropagated()
	metrics["broker_server"]["sync_requests_received"] = server.CheckSyncRequestsReceived()
	metrics["broker_server"]["sync_requests_propagated"] = server.CheckSyncRequestsPropagated()
	metrics["broker_server"]["connection_count"] = uint64(len(server.connectionAsyncSubscriptions))
	metrics["broker_server"]["async_topic_count"] = uint64(len(server.asyncTopicSubscriptions))
	metrics["broker_server"]["sync_topic_count"] = uint64(len(server.syncTopicSubscriptions))
	server.mutex.Unlock()
	for systemgeName, systemgeMetrics := range server.systemgeServer.CheckMetrics() {
		metrics[systemgeName] = systemgeMetrics
	}
	metrics["message_handler"] = server.messageHandler.CheckMetrics()
	return metrics
}
func (server *Server) GetMetrics() map[string]map[string]uint64 {
	metrics := map[string]map[string]uint64{}
	metrics["broker_server"] = map[string]uint64{}
	server.mutex.Lock()
	metrics["broker_server"]["async_messages_received"] = server.GetAsyncMessagesReceived()
	metrics["broker_server"]["async_messages_propagated"] = server.GetAsyncMessagesPropagated()
	metrics["broker_server"]["sync_requests_received"] = server.GetSyncRequestsReceived()
	metrics["broker_server"]["sync_requests_propagated"] = server.GetSyncRequestsPropagated()
	metrics["broker_server"]["connection_count"] = uint64(len(server.connectionAsyncSubscriptions))
	metrics["broker_server"]["async_topic_count"] = uint64(len(server.asyncTopicSubscriptions))
	metrics["broker_server"]["sync_topic_count"] = uint64(len(server.syncTopicSubscriptions))
	server.mutex.Unlock()
	for systemgeName, systemgeMetrics := range server.systemgeServer.GetMetrics() {
		metrics[systemgeName] = systemgeMetrics
	}
	metrics["message_handler"] = server.messageHandler.GetMetrics()
	return metrics
}

func (server *Server) CheckAsyncMessagesReceived() uint64 {
	return server.asyncMessagesReceived.Load()
}
func (server *Server) GetAsyncMessagesReceived() uint64 {
	return server.asyncMessagesReceived.Swap(0)
}

func (server *Server) CheckAsyncMessagesPropagated() uint64 {
	return server.asyncMessagesPropagated.Load()
}
func (server *Server) GetAsyncMessagesPropagated() uint64 {
	return server.asyncMessagesPropagated.Swap(0)
}

func (server *Server) CheckSyncRequestsReceived() uint64 {
	return server.syncRequestsReceived.Load()
}
func (server *Server) GetSyncRequestsReceived() uint64 {
	return server.syncRequestsReceived.Swap(0)
}

func (server *Server) CheckSyncRequestsPropagated() uint64 {
	return server.syncRequestsPropagated.Load()
}
func (server *Server) GetSyncRequestsPropagated() uint64 {
	return server.syncRequestsPropagated.Swap(0)
}
