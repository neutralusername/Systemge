package BrokerServer

import "github.com/neutralusername/Systemge/DashboardHelpers"

func (server *Server) CheckMetrics() map[string]map[string]uint64 {
	server.mutex.Lock()
	metrics := map[string]map[string]uint64{
		"broker_server": {
			"async_messages_received":   server.CheckAsyncMessagesReceived(),
			"async_messages_propagated": server.CheckAsyncMessagesPropagated(),
			"sync_requests_received":    server.CheckSyncRequestsReceived(),
			"sync_requests_propagated":  server.CheckSyncRequestsPropagated(),
			"connection_count":          uint64(len(server.connectionAsyncSubscriptions)),
			"async_topic_count":         uint64(len(server.asyncTopicSubscriptions)),
			"sync_topic_count":          uint64(len(server.syncTopicSubscriptions)),
		},
	}
	server.mutex.Unlock()
	DashboardHelpers.MergeMetrics(metrics, server.systemgeServer.CheckMetrics())
	DashboardHelpers.MergeMetrics(metrics, server.messageHandler.CheckMetrics())
	return metrics
}
func (server *Server) GetMetrics() map[string]map[string]uint64 {
	server.mutex.Lock()
	metrics := map[string]map[string]uint64{
		"broker_server": {
			"async_messages_received":   server.GetAsyncMessagesReceived(),
			"async_messages_propagated": server.GetAsyncMessagesPropagated(),
			"sync_requests_received":    server.GetSyncRequestsReceived(),
			"sync_requests_propagated":  server.GetSyncRequestsPropagated(),
			"connection_count":          uint64(len(server.connectionAsyncSubscriptions)),
			"async_topic_count":         uint64(len(server.asyncTopicSubscriptions)),
			"sync_topic_count":          uint64(len(server.syncTopicSubscriptions)),
		},
	}
	server.mutex.Unlock()
	DashboardHelpers.MergeMetrics(metrics, server.systemgeServer.GetMetrics())
	DashboardHelpers.MergeMetrics(metrics, server.messageHandler.GetMetrics())
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
