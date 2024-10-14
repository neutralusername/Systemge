package ChannelConnection

import "github.com/neutralusername/Systemge/Metrics"

func (client *ChannelConnection[T]) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"bytesSent":        client.BytesSent.Swap(0),
			"bytesReceived":    client.BytesReceived.Swap(0),
			"messagesSent":     client.MessagesSent.Swap(0),
			"messagesReceived": client.MessagesReceived.Swap(0),
		}),
	}
}

func (client *ChannelConnection[T]) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"bytesSent":        client.BytesSent.Load(),
			"bytesReceived":    client.BytesReceived.Load(),
			"messagesSent":     client.MessagesSent.Load(),
			"messagesReceived": client.MessagesReceived.Load(),
		}),
	}
}
