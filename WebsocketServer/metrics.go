package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *WebsocketServer) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     server.CheckWebsocketConnectionMessagesBytesSent(),
			"bytes_received": server.CheckWebsocketConnectionMessagesBytesReceived(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_messageTransmissions", Metrics.New(
		map[string]uint64{
			"incoming_messages": server.CheckWebsocketConnectionMessagesReceived(),
			"outgoing_messages": server.CheckWebsocketConnectionMessagesSent(),
			"failed_messages":   server.CheckWebsocketConnectionMessagesFailed(),
		}),
	)

	metricsTypes.AddMetrics("websocketServer_connections", Metrics.New(
		map[string]uint64{
			"connections_accepted": server.CheckWebsocketConnectionsAccepted(),
			"connections_failed":   server.CheckWebsocketConnectionsFailed(),
			"connections_rejected": server.CheckWebsocketConnectionsRejected(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_clientSessionManager", Metrics.New(
		map[string]uint64{
			// todo
		}),
	)
	metricsTypes.AddMetrics("websocketServer_groupSessionManager", Metrics.New(
		map[string]uint64{
			// todo
		}),
	)
	metricsTypes.Merge(server.httpServer.CheckMetrics())
	return metricsTypes
}

func (server *WebsocketServer) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     uint64(server.GetWebsocketConnectionMessagesBytesSent()),
			"bytes_received": uint64(server.GetWebsocketConnectionMessagesBytesReceived()),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_messageTransmissions", Metrics.New(
		map[string]uint64{
			"incoming_messages": uint64(server.GetWebsocketConnectionMessagesReceived()),
			"outgoing_messages": uint64(server.GetWebsocketConnectionMessagesSent()),
			"failed_messages":   uint64(server.GetWebsocketConnectionMessagesFailed()),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_connections", Metrics.New(
		map[string]uint64{
			"connections_accepted": server.GetWebsocketConnectionsAccepted(),
			"connections_failed":   server.GetWebsocketConnectionsFailed(),
			"connections_rejected": server.GetWebsocketConnectionsRejected(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_clientSessionManager", Metrics.New(
		map[string]uint64{
			// todo
		}),
	)
	metricsTypes.AddMetrics("websocketServer_groupSessionManager", Metrics.New(
		map[string]uint64{
			// todo
		}),
	)
	metricsTypes.Merge(server.httpServer.GetMetrics())
	return metricsTypes
}

func (server *WebsocketServer) GetWebsocketConnectionsAccepted() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionsAccepted() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionsFailed() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionsFailed() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionsRejected() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionsRejected() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesReceived() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesReceived() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesSent() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesSent() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesFailed() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesFailed() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesBytesReceived() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesBytesReceived() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesBytesSent() uint64 {
	return server.websocketConnectionMessagesBytesSent.Swap(0)
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesBytesSent() uint64 {
	return server.websocketConnectionMessagesBytesSent.Load()
}
