package ConnectionWebsocket

import "github.com/neutralusername/Systemge/Tools"

func (connection *WebsocketConnection) GetMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketClient": Tools.NewMetrics(map[string]uint64{
			"bytesSent":        connection.BytesSent.Swap(0),
			"bytesReceived":    connection.BytesReceived.Swap(0),
			"messagesSent":     connection.MessagesSent.Swap(0),
			"messagesReceived": connection.MessagesReceived.Swap(0),
		}),
	}
}

func (connection *WebsocketConnection) CheckMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketClient": Tools.NewMetrics(map[string]uint64{
			"bytesSent":        connection.BytesSent.Load(),
			"bytesReceived":    connection.BytesReceived.Load(),
			"messagesSent":     connection.MessagesSent.Load(),
			"messagesReceived": connection.MessagesReceived.Load(),
		}),
	}
}
