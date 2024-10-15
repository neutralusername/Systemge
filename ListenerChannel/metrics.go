package ListenerChannel

import "github.com/neutralusername/Systemge/Metrics"

func (listener *ChannelListener[T]) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Swap(0),
			"clientsFailed":      listener.ClientsFailed.Swap(0),
		}),
	}
}

func (listener *ChannelListener[T]) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Load(),
			"clientsFailed":      listener.ClientsFailed.Load(),
		}),
	}
}
