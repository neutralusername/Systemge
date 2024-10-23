package connectionChannel

import (
	"github.com/neutralusername/systemge/tools"
)

func (connection *ChannelConnection[T]) GetMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketClient": tools.NewMetrics(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Swap(0),
			"messagesReceived": connection.MessagesReceived.Swap(0),
		}),
	}
}

func (connection *ChannelConnection[T]) CheckMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketClient": tools.NewMetrics(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Load(),
			"messagesReceived": connection.MessagesReceived.Load(),
		}),
	}
}
