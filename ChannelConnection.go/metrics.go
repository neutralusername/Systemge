package ChannelConnection

import "github.com/neutralusername/Systemge/Metrics"

func (client *ChannelConnection[T]) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"messagesSent":     client.MessagesSent.Swap(0),
			"messagesReceived": client.MessagesReceived.Swap(0),
		}),
	}
}

func (client *ChannelConnection[T]) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"messagesSent":     client.MessagesSent.Load(),
			"messagesReceived": client.MessagesReceived.Load(),
		}),
	}
}
