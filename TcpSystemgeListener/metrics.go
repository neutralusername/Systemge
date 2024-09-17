package TcpSystemgeListener

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (listener *TcpSystemgeListener) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", Metrics.New(
		map[string]uint64{
			"connection_attempts":  listener.CheckConnectionAttempts(),
			"failed_connections":   listener.CheckFailedConnections(),
			"rejected_connections": listener.CheckRejectedConnections(),
			"accepted_connections": listener.CheckAcceptedConnections(),
		}),
	)
	return metricsTypes
}
func (listener *TcpSystemgeListener) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", Metrics.New(
		map[string]uint64{
			"connection_attempts":  listener.GetConnectionAttempts(),
			"failed_connections":   listener.GetFailedConnections(),
			"rejected_connections": listener.GetRejectedConnections(),
			"accepted_connections": listener.GetAcceptedConnections(),
		}),
	)
	return metricsTypes
}

func (listener *TcpSystemgeListener) CheckConnectionAttempts() uint64 {
	return listener.connectionAttempts.Load()
}
func (listener *TcpSystemgeListener) GetConnectionAttempts() uint64 {
	return listener.connectionAttempts.Swap(0)
}

func (listener *TcpSystemgeListener) CheckFailedConnections() uint64 {
	return listener.failedConnections.Load()
}
func (listener *TcpSystemgeListener) GetFailedConnections() uint64 {
	return listener.failedConnections.Swap(0)
}

func (listener *TcpSystemgeListener) CheckRejectedConnections() uint64 {
	return listener.rejectedConnections.Load()
}
func (listener *TcpSystemgeListener) GetRejectedConnections() uint64 {
	return listener.rejectedConnections.Swap(0)
}

func (listener *TcpSystemgeListener) CheckAcceptedConnections() uint64 {
	return listener.acceptedConnections.Load()
}
func (listener *TcpSystemgeListener) GetAcceptedConnections() uint64 {
	return listener.acceptedConnections.Swap(0)
}
