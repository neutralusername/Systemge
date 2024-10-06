package WebsocketServer

import "github.com/neutralusername/Systemge/Metrics"

func (server *WebsocketServer) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	/* metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     server.CheckWebsocketClientMessagesBytesSent(),
			"bytes_received": server.CheckWebsocketClientMessagesBytesReceived(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_messageTransmissions", Metrics.New(
		map[string]uint64{
			"messages_received": server.CheckWebsocketClientMessagesReceived(),
			"messages_sent":     server.CheckWebsocketClientMessagesSent(),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_invalidMessages", Metrics.New(
		map[string]uint64{
			"invalid_messages_received":  server.CheckWebsocketClientInvalidMessagesReceived(),
			"rejected_messages_received": server.CheckWebsocketClientRejectedMessagesReceived(),
		}),
	)
	metricsTypes.Merge(server.websocketListener.CheckMetrics()) */
	return metricsTypes
}

func (server *WebsocketServer) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	/* metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
		map[string]uint64{
			"bytes_sent":     uint64(server.GetWebsocketClientMessagesBytesSent()),
			"bytes_received": uint64(server.GetWebsocketClientMessagesBytesReceived()),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_messageTransmissions", Metrics.New(
		map[string]uint64{
			"messages_received": uint64(server.GetWebsocketClientMessagesReceived()),
			"messages_sent":     uint64(server.GetWebsocketClientMessagesSent()),
		}),
	)
	metricsTypes.AddMetrics("websocketServer_invalidMessages", Metrics.New(
		map[string]uint64{
			"invalid_messages_received":  uint64(server.GetWebsocketClientInvalidMessagesReceived()),
			"rejected_messages_received": uint64(server.GetWebsocketClientRejectedMessagesReceived()),
		}),
	)
	metricsTypes.AddMetrics("websocketListener", Metrics.New(
		map[string]uint64{
			"connection_accepted": server.GetClientsAccepted(),
			"clients_failed":      server.GetClientsFailed(),
			"clients_rejected":    server.GetClientsRejected(),
		},
	)) */
	return metricsTypes
}
