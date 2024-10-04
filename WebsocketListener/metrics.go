package WebsocketListener

import "github.com/neutralusername/Systemge/Metrics"

func (listener *WebsocketListener) GetConnectionsAccepted() uint64 {
	return listener.connectionsAccepted.Swap(0)
}
func (listener *WebsocketListener) CheckConnectionsAccepted() uint64 {
	return listener.connectionsAccepted.Load()
}

func (listener *WebsocketListener) GetConnectionsFailed() uint64 {
	return listener.connectionsFailed.Swap(0)
}
func (listener *WebsocketListener) CheckConnectionsFailed() uint64 {
	return listener.connectionsFailed.Load()
}

func (listener *WebsocketListener) GetConnectionsRejected() uint64 {
	return listener.connectionsRejected.Swap(0)
}
func (listener *WebsocketListener) CheckConnectionsRejected() uint64 {
	return listener.connectionsRejected.Load()
}

func (listener *WebsocketListener) GetMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted":  listener.GetConnectionsAccepted(),
			"connectionsFailed":   listener.GetConnectionsFailed(),
			"connectionsRejected": listener.GetConnectionsRejected(),
		}),
	}
}

func (listener *WebsocketListener) CheckMetrics() Metrics.MetricsTypes {
	return Metrics.MetricsTypes{
		"websocketListener": Metrics.New(map[string]uint64{
			"connectionAccepted":  listener.CheckConnectionsAccepted(),
			"connectionsFailed":   listener.CheckConnectionsFailed(),
			"connectionsRejected": listener.CheckConnectionsRejected(),
		}),
	}
}
