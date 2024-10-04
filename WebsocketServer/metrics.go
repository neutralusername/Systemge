package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *WebsocketServer) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
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
	metricsTypes.Merge(server.websocketListener.CheckMetrics())
	return metricsTypes
}

func (server *WebsocketServer) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("websocketServer_byteTransmissions", Metrics.New(
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
			"connection_accepted":  server.GetConnectionAccepted(),
			"connections_failed":   server.GetConnectionsFailed(),
			"connections_rejected": server.GetConnectionsRejected(),
		},
	))
	return metricsTypes
}

func (server *WebsocketServer) GetConnectionAccepted() uint64 {
	return server.websocketListener.GetConnectionsAccepted()
}
func (server *WebsocketServer) CheckConnectionAccepted() uint64 {
	return server.websocketListener.CheckConnectionsAccepted()
}

func (server *WebsocketServer) GetConnectionsFailed() uint64 {
	return server.websocketListener.GetConnectionsFailed()
}
func (server *WebsocketServer) CheckConnectionsFailed() uint64 {
	return server.websocketListener.CheckConnectionsFailed()
}

func (server *WebsocketServer) GetConnectionsRejected() uint64 {
	return server.websocketListener.GetConnectionsRejected()
}
func (server *WebsocketServer) CheckConnectionsRejected() uint64 {
	return server.websocketListener.CheckConnectionsRejected()
}

func (server *WebsocketServer) GetWebsocketClientMessagesBytesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesSent uint64
	for _, websocketClient := range websocketClients {
		bytesSent += websocketClient.GetBytesSent()
	}
	return bytesSent
}
func (server *WebsocketServer) CheckWebsocketClientMessagesBytesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesSent uint64
	for _, websocketClient := range websocketClients {
		bytesSent += websocketClient.CheckBytesSent()
	}
	return bytesSent
}

func (server *WebsocketServer) GetWebsocketClientMessagesBytesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesReceived uint64
	for _, websocketClient := range websocketClients {
		bytesReceived += websocketClient.GetBytesReceived()
	}
	return bytesReceived
}
func (server *WebsocketServer) CheckWebsocketClientMessagesBytesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesReceived uint64
	for _, websocketClient := range websocketClients {
		bytesReceived += websocketClient.CheckBytesReceived()
	}
	return bytesReceived
}

func (server *WebsocketServer) GetWebsocketClientMessagesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesSent uint64
	for _, websocketClient := range websocketClients {
		messagesSent += websocketClient.GetMessagesSent()
	}
	return messagesSent
}
func (server *WebsocketServer) CheckWebsocketClientMessagesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesSent uint64
	for _, websocketClient := range websocketClients {
		messagesSent += websocketClient.CheckMessagesSent()
	}
	return messagesSent
}

func (server *WebsocketServer) GetWebsocketClientMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesReceived uint64
	for _, websocketClient := range websocketClients {
		messagesReceived += websocketClient.GetMessagesReceived()
	}
	return messagesReceived
}
func (server *WebsocketServer) CheckWebsocketClientMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesReceived uint64
	for _, websocketClient := range websocketClients {
		messagesReceived += websocketClient.CheckMessagesReceived()
	}
	return messagesReceived
}

func (server *WebsocketServer) GetWebsocketClientInvalidMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var invalidMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		invalidMessagesReceived += websocketClient.GetInvalidMessagesReceived()
	}
	return invalidMessagesReceived
}
func (server *WebsocketServer) CheckWebsocketClientInvalidMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var invalidMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		invalidMessagesReceived += websocketClient.CheckInvalidMessagesReceived()
	}
	return invalidMessagesReceived
}

func (server *WebsocketServer) GetWebsocketClientRejectedMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var rejectedMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		rejectedMessagesReceived += websocketClient.GetRejectedMessagesReceived()
	}
	return rejectedMessagesReceived
}
func (server *WebsocketServer) CheckWebsocketClientRejectedMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var rejectedMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		rejectedMessagesReceived += websocketClient.CheckRejectedMessagesReceived()
	}
	return rejectedMessagesReceived
}
