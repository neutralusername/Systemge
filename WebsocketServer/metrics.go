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
			"messages_received": server.CheckWebsocketConnectionMessagesReceived(),
			"messages_sent":     server.CheckWebsocketConnectionMessagesSent(),
		}),
	)
	metricsTypes.Merge(server.websocketListener.CheckMetrics())
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
			"messages_received": uint64(server.GetWebsocketConnectionMessagesReceived()),
			"messages_sent":     uint64(server.GetWebsocketConnectionMessagesSent()),
		}),
	)
	metricsTypes.Merge(server.websocketListener.GetMetrics())
	return metricsTypes
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesBytesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesSent uint64
	for _, websocketClient := range websocketClients {
		bytesSent += websocketClient.GetBytesSent()
	}
	return bytesSent
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesBytesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesSent uint64
	for _, websocketClient := range websocketClients {
		bytesSent += websocketClient.CheckBytesSent()
	}
	return bytesSent
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesBytesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesReceived uint64
	for _, websocketClient := range websocketClients {
		bytesReceived += websocketClient.GetBytesReceived()
	}
	return bytesReceived
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesBytesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var bytesReceived uint64
	for _, websocketClient := range websocketClients {
		bytesReceived += websocketClient.CheckBytesReceived()
	}
	return bytesReceived
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesSent uint64
	for _, websocketClient := range websocketClients {
		messagesSent += websocketClient.GetMessagesSent()
	}
	return messagesSent
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesSent() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesSent uint64
	for _, websocketClient := range websocketClients {
		messagesSent += websocketClient.CheckMessagesSent()
	}
	return messagesSent
}

func (server *WebsocketServer) GetWebsocketConnectionMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesReceived uint64
	for _, websocketClient := range websocketClients {
		messagesReceived += websocketClient.GetMessagesReceived()
	}
	return messagesReceived
}
func (server *WebsocketServer) CheckWebsocketConnectionMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var messagesReceived uint64
	for _, websocketClient := range websocketClients {
		messagesReceived += websocketClient.CheckMessagesReceived()
	}
	return messagesReceived
}

func (server *WebsocketServer) GetWebsocketConnectionInvalidMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var invalidMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		invalidMessagesReceived += websocketClient.GetInvalidMessagesReceived()
	}
	return invalidMessagesReceived
}
func (server *WebsocketServer) CheckWebsocketConnectionInvalidMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var invalidMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		invalidMessagesReceived += websocketClient.CheckInvalidMessagesReceived()
	}
	return invalidMessagesReceived
}

func (server *WebsocketServer) GetWebsocketConnectionRejectedMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var rejectedMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		rejectedMessagesReceived += websocketClient.GetRejectedMessagesReceived()
	}
	return rejectedMessagesReceived
}
func (server *WebsocketServer) CheckWebsocketConnectionRejectedMessagesReceived() uint64 {
	websocketClients := server.GetWebsocketClients()
	var rejectedMessagesReceived uint64
	for _, websocketClient := range websocketClients {
		rejectedMessagesReceived += websocketClient.CheckRejectedMessagesReceived()
	}
	return rejectedMessagesReceived
}
