package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *WebsocketServer) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     server.CheckBytesSentCounter(),
			"bytes_received": server.CheckBytesReceivedCounter(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_messageTransmissions", Metrics.New(
		map[string]uint64{
			"incoming_messages": uint64(server.CheckIncomingMessageCounter()),
			"outgoing_messages": uint64(server.CheckOutgoingMessageCounter()),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_stats", Metrics.New(
		map[string]uint64{
			"active_connections": uint64(server.GetClientCount()),
		}),
	)
	metricsTypes.Merge(server.httpServer.CheckMetrics())
	return metricsTypes
}

func (server *WebsocketServer) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     server.GetBytesSentCounter(),
			"bytes_received": server.GetBytesReceivedCounter(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_messageTransmissions", Metrics.New(
		map[string]uint64{
			"incoming_messages": uint64(server.GetIncomingMessageCounter()),
			"outgoing_messages": uint64(server.GetOutgoingMessageCounter()),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_stats", Metrics.New(
		map[string]uint64{
			"active_connections": uint64(server.GetClientCount()),
		}),
	)
	metricsTypes.Merge(server.httpServer.GetMetrics())
	return metricsTypes
}

func (server *WebsocketServer) GetBytesSentCounter() uint64 {
	return server.bytesSentCounter.Swap(0)
}
func (server *WebsocketServer) CheckBytesSentCounter() uint64 {
	return server.bytesSentCounter.Load()
}

func (server *WebsocketServer) GetBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Swap(0)
}
func (server *WebsocketServer) CheckBytesReceivedCounter() uint64 {
	return server.bytesReceivedCounter.Load()
}

func (server *WebsocketServer) GetIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Swap(0)
}
func (server *WebsocketServer) CheckIncomingMessageCounter() uint32 {
	return server.incomingMessageCounter.Load()
}

func (server *WebsocketServer) GetOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Swap(0)
}
func (server *WebsocketServer) CheckOutgoingMessageCounter() uint32 {
	return server.outgoigMessageCounter.Load()
}
