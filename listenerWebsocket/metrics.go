package listenerWebsocket

import (
	"github.com/neutralusername/Systemge/tools"
)

func (listener *WebsocketListener) GetMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketListener": tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Swap(0),
			"clientsFailed":      listener.ClientsFailed.Swap(0),
			"clientsRejected":    listener.ClientsRejected.Swap(0),
		}),
	}
}

func (listener *WebsocketListener) CheckMetrics() tools.MetricsTypes {
	return tools.MetricsTypes{
		"websocketListener": tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Load(),
			"clientsFailed":      listener.ClientsFailed.Load(),
			"clientsRejected":    listener.ClientsRejected.Load(),
		}),
	}
}
