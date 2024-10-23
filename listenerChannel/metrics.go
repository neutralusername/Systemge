package listenerChannel

import (
	"github.com/neutralusername/systemge/tools"
)

func (listener *ChannelListener[T]) GetMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketListener": tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Swap(0),
			"clientsFailed":      listener.ClientsFailed.Swap(0),
		}),
	}
}

func (listener *ChannelListener[T]) CheckMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketListener": tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Load(),
			"clientsFailed":      listener.ClientsFailed.Load(),
		}),
	}
}
