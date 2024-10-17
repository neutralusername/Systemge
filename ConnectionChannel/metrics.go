package ConnectionChannel

import (
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *ChannelConnection[T]) GetMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketClient": Tools.NewMetrics(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Swap(0),
			"messagesReceived": connection.MessagesReceived.Swap(0),
		}),
	}
}

func (connection *ChannelConnection[T]) CheckMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketClient": Tools.NewMetrics(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Load(),
			"messagesReceived": connection.MessagesReceived.Load(),
		}),
	}
}
