package TcpSystemgeListener

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (listener *TcpSystemgeListener) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", Metrics.New(
		map[string]uint64{
			"total_connection_attempts":    listener.CheckConnectionAttempts(),
			"failed_connection_attempts":   listener.CheckFailedConnectionAttempts(),
			"rejected_connection_attempts": listener.CheckRejectedConnectionAttempts(),
			"accepted_connection_attempts": listener.CheckAcceptedConnectionAttempts(),
		}),
	)
	return metricsTypes
}
func (listener *TcpSystemgeListener) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", Metrics.New(
		map[string]uint64{
			"total_connection_attempts":    listener.GetConnectionAttempts(),
			"failed_connection_attempts":   listener.GetFailedConnectionAttempts(),
			"rejected_connection_attempts": listener.GetRejectedConnectionAttempts(),
			"accepted_connection_attempts": listener.GetAcceptedConnectionAttempts(),
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

func (listener *TcpSystemgeListener) CheckFailedConnectionAttempts() uint64 {
	return listener.failedConnectionAttempts.Load()
}
func (listener *TcpSystemgeListener) GetFailedConnectionAttempts() uint64 {
	return listener.failedConnectionAttempts.Swap(0)
}

func (listener *TcpSystemgeListener) CheckRejectedConnectionAttempts() uint64 {
	return listener.rejectedConnectionAttempts.Load()
}
func (listener *TcpSystemgeListener) GetRejectedConnectionAttempts() uint64 {
	return listener.rejectedConnectionAttempts.Swap(0)
}

func (listener *TcpSystemgeListener) CheckAcceptedConnectionAttempts() uint64 {
	return listener.acceptedConnectionAttempts.Load()
}
func (listener *TcpSystemgeListener) GetAcceptedConnectionAttempts() uint64 {
	return listener.acceptedConnectionAttempts.Swap(0)
}
