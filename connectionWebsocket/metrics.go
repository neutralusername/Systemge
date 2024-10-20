package connectionWebsocket

import "github.com/neutralusername/systemge/tools"

func (connection *WebsocketConnection) GetMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketClient": tools.NewMetrics(map[string]uint64{
			"bytesSent":        connection.BytesSent.Swap(0),
			"bytesReceived":    connection.BytesReceived.Swap(0),
			"messagesSent":     connection.MessagesSent.Swap(0),
			"messagesReceived": connection.MessagesReceived.Swap(0),
		}),
	}
}

func (connection *WebsocketConnection) CheckMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketClient": tools.NewMetrics(map[string]uint64{
			"bytesSent":        connection.BytesSent.Load(),
			"bytesReceived":    connection.BytesReceived.Load(),
			"messagesSent":     connection.MessagesSent.Load(),
			"messagesReceived": connection.MessagesReceived.Load(),
		}),
	}
}
