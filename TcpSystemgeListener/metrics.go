package TcpSystemgeListener

import (
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

func (listener *TcpSystemgeListener) CheckMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"tcp_systemge_listener": {
			KeyValuePairs: map[string]uint64{
				"connection_attempts":  listener.CheckConnectionAttempts(),
				"failed_connections":   listener.CheckFailedConnections(),
				"rejected_connections": listener.CheckRejectedConnections(),
				"accepted_connections": listener.CheckAcceptedConnections(),
			},
			Time: time.Now(),
		},
	}
}
func (listener *TcpSystemgeListener) GetMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"tcp_systemge_listener": {
			KeyValuePairs: map[string]uint64{
				"connection_attempts":  listener.GetConnectionAttempts(),
				"failed_connections":   listener.GetFailedConnections(),
				"rejected_connections": listener.GetRejectedConnections(),
				"accepted_connections": listener.GetAcceptedConnections(),
			},
			Time: time.Now(),
		},
	}
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
