package WebsocketListener

import "github.com/neutralusername/Systemge/Metrics"

func (listener *WebsocketListener) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Swap(0),
			"clientsFailed":      listener.ClientsFailed.Swap(0),
			"clientsRejected":    listener.ClientsRejected.Swap(0),
		}),
	}
}

func (listener *WebsocketListener) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Load(),
			"clientsFailed":      listener.ClientsFailed.Load(),
			"clientsRejected":    listener.ClientsRejected.Load(),
		}),
	}
}
