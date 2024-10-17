package ListenerTcp

import "github.com/neutralusername/Systemge/Tools"

func (listener *TcpListener) CheckMetrics() Tools.MetricsTypes {
	metricsTypes := Tools.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", Tools.NewMetrics(
		map[string]uint64{
			/* 	"total_connection_attempts":    listener.CheckConnectionAttempts(),
			"failed_connection_attempts":   listener.CheckFailedConnectionAttempts(),
			"rejected_connection_attempts": listener.CheckRejectedConnectionAttempts(),
			"accepted_connection_attempts": listener.CheckAcceptedConnectionAttempts(), */
		}),
	)
	return metricsTypes
}
func (listener *TcpListener) GetMetrics() Tools.MetricsTypes {
	metricsTypes := Tools.NewMetricsTypes()
	metricsTypes.AddMetrics("tcp_systemge_listener", Tools.NewMetrics(
		map[string]uint64{
			/* 		"total_connection_attempts":    listener.GetConnectionAttempts(),
			"failed_connection_attempts":   listener.GetFailedConnectionAttempts(),
			"rejected_connection_attempts": listener.GetRejectedConnectionAttempts(),
			"accepted_connection_attempts": listener.GetAcceptedConnectionAttempts(), */
		}),
	)
	return metricsTypes
}
