package ListenerChannel

import (
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *ChannelListener[T]) GetMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketListener": Tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Swap(0),
			"clientsFailed":      listener.ClientsFailed.Swap(0),
		}),
	}
}

func (listener *ChannelListener[T]) CheckMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketListener": Tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Load(),
			"clientsFailed":      listener.ClientsFailed.Load(),
		}),
	}
}
