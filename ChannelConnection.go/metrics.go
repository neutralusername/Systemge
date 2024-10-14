package ChannelConnection

import "github.com/neutralusername/Systemge/Metrics"

func (connection *ChannelConnection[T]) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Swap(0),
			"messagesReceived": connection.MessagesReceived.Swap(0),
		}),
	}
}

func (connection *ChannelConnection[T]) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketClient": Metrics.New(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Load(),
			"messagesReceived": connection.MessagesReceived.Load(),
		}),
	}
}
