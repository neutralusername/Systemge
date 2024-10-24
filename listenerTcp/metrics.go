package listenerTcp

import "github.com/neutralusername/systemge/tools"

func (listener *TcpListener) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", tools.NewMetrics(
		map[string]uint64{
			/* 	"total_connection_attempts":    listener.CheckConnectionAttempts(),
			"failed_connection_attempts":   listener.CheckFailedConnectionAttempts(),
			"rejected_connection_attempts": listener.CheckRejectedConnectionAttempts(),
			"accepted_connection_attempts": listener.CheckAcceptedConnectionAttempts(), */
		}),
	)
	return metricsTypes
}
func (listener *TcpListener) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", tools.NewMetrics(
		map[string]uint64{
			/* 		"total_connection_attempts":    listener.GetConnectionAttempts(),
			"failed_connection_attempts":   listener.GetFailedConnectionAttempts(),
			"rejected_connection_attempts": listener.GetRejectedConnectionAttempts(),
			"accepted_connection_attempts": listener.GetAcceptedConnectionAttempts(), */
		}),
	)
	return metricsTypes
}
