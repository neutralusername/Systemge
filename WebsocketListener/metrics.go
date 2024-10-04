package WebsocketListener

import "github.com/neutralusername/Systemge/Metrics"

func (listener *WebsocketListener) GetClientsAccepted() uint64 {
	return listener.clientsAccepted.Swap(0)
}
func (listener *WebsocketListener) CheckClientsAccepted() uint64 {
	return listener.clientsAccepted.Load()
}

func (listener *WebsocketListener) GetClientsFailed() uint64 {
	return listener.clientsFailed.Swap(0)
}
func (listener *WebsocketListener) CheckClientsFailed() uint64 {
	return listener.clientsFailed.Load()
}

func (listener *WebsocketListener) GetClientsRejected() uint64 {
	return listener.clientsRejected.Swap(0)
}
func (listener *WebsocketListener) CheckClientsRejected() uint64 {
	return listener.clientsRejected.Load()
}

func (listener *WebsocketListener) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted": listener.GetClientsAccepted(),
			"clientsFailed":      listener.GetClientsFailed(),
			"clientsRejected":    listener.GetClientsRejected(),
		}),
	}
}

func (listener *WebsocketListener) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted": listener.CheckClientsAccepted(),
			"clientsFailed":      listener.CheckClientsFailed(),
			"clientsRejected":    listener.CheckClientsRejected(),
		}),
	}
}
