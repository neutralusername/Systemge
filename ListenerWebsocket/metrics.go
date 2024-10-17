package ListenerWebsocket

import (
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *WebsocketListener) GetMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketListener": Tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Swap(0),
			"clientsFailed":      listener.ClientsFailed.Swap(0),
			"clientsRejected":    listener.ClientsRejected.Swap(0),
		}),
	}
}

func (listener *WebsocketListener) CheckMetrics() Tools.MetricsTypes {
	return Tools.MetricsTypes{
		"websocketListener": Tools.NewMetrics(map[string]uint64{
			"connectionAccepted": listener.ClientsAccepted.Load(),
			"clientsFailed":      listener.ClientsFailed.Load(),
			"clientsRejected":    listener.ClientsRejected.Load(),
		}),
	}
}
