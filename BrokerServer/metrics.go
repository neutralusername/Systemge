package BrokerServer

import (
	"github.com/neutralusername/systemge/Metrics"
)

func (server *Server) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	server.mutex.Lock()
	metricsTypes.AddMetrics("brokerServer_messagesPropagated", Metrics.New(
		map[string]uint64{
			"async_messages_propagated": server.CheckAsyncMessagesPropagated(),
			"sync_requests_propagated":  server.CheckSyncRequestsPropagated(),
		},
	))
	metricsTypes.AddMetrics("brokerServer_stats", Metrics.New(
		map[string]uint64{
			"connection_count":  uint64(len(server.connectionAsyncSubscriptions)),
			"async_topic_count": uint64(len(server.asyncTopicSubscriptions)),
			"sync_topic_count":  uint64(len(server.syncTopicSubscriptions)),
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
	metricsTypes.AddMetrics("brokerServer_messagesPropagated", Metrics.New(
		map[string]uint64{
			"async_messages_propagated": server.GetAsyncMessagesPropagated(),
			"sync_requests_propagated":  server.GetSyncRequestsPropagated(),
		},
	))
	metricsTypes.AddMetrics("brokerServer_stats", Metrics.New(
		map[string]uint64{
			"connection_count":  uint64(len(server.connectionAsyncSubscriptions)),
			"async_topic_count": uint64(len(server.asyncTopicSubscriptions)),
			"sync_topic_count":  uint64(len(server.syncTopicSubscriptions)),
		},
	))
	server.mutex.Unlock()

	metricsTypes.Merge(server.systemgeServer.GetMetrics())
	metricsTypes.Merge(server.messageHandler.GetMetrics())
	return metricsTypes
}

func (server *Server) CheckAsyncMessagesPropagated() uint64 {
	return server.asyncMessagesPropagated.Load()
}
func (server *Server) GetAsyncMessagesPropagated() uint64 {
	return server.asyncMessagesPropagated.Swap(0)
}

func (server *Server) CheckSyncRequestsPropagated() uint64 {
	return server.syncRequestsPropagated.Load()
}
func (server *Server) GetSyncRequestsPropagated() uint64 {
	return server.syncRequestsPropagated.Swap(0)
}
