package BrokerServer

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *Server) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	server.mutex.Lock()
	metricsTypes.AddMetrics("broker_server", Metrics.New(
		map[string]uint64{
			"async_messages_received":   server.CheckAsyncMessagesReceived(),
			"async_messages_propagated": server.CheckAsyncMessagesPropagated(),
			"sync_requests_received":    server.CheckSyncRequestsReceived(),
			"sync_requests_propagated":  server.CheckSyncRequestsPropagated(),
			"connection_count":          uint64(len(server.connectionAsyncSubscriptions)),
			"async_topic_count":         uint64(len(server.asyncTopicSubscriptions)),
			"sync_topic_count":          uint64(len(server.syncTopicSubscriptions)),
		},
	))
	server.mutex.Unlock()

	metricsTypes.Merge(server.systemgeServer.CheckMetrics())
	metricsTypes.Merge(server.messageHandler.CheckMetrics())
	return metricsTypes
}
func (server *Server) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	server.mutex.Lock()
	metricsTypes.AddMetrics("broker_server", Metrics.New(
		map[string]uint64{
			"async_messages_received":   server.GetAsyncMessagesReceived(),
			"async_messages_propagated": server.GetAsyncMessagesPropagated(),
			"sync_requests_received":    server.GetSyncRequestsReceived(),
			"sync_requests_propagated":  server.GetSyncRequestsPropagated(),
			"connection_count":          uint64(len(server.connectionAsyncSubscriptions)),
			"async_topic_count":         uint64(len(server.asyncTopicSubscriptions)),
			"sync_topic_count":          uint64(len(server.syncTopicSubscriptions)),
		},
	))
	server.mutex.Unlock()

	metricsTypes.Merge(server.systemgeServer.GetMetrics())
	metricsTypes.Merge(server.messageHandler.GetMetrics())
	return metricsTypes
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
