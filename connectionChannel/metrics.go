package connectionChannel

import (
	"github.com/neutralusername/systemge/tools"
)

func (connection *ChannelConnection[D]) GetMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketClient": tools.NewMetrics(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Swap(0),
			"messagesReceived": connection.MessagesReceived.Swap(0),
		}),
	}
}

func (connection *ChannelConnection[D]) CheckMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketClient": tools.NewMetrics(map[string]uint64{
			"messagesSent":     connection.MessagesSent.Load(),
			"messagesReceived": connection.MessagesReceived.Load(),
		}),
	}
}
