package TcpSystemgeListener

import (
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

func (listener *TcpListener) CheckMetrics() map[string]*Metrics.Metrics {
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
func (listener *TcpListener) GetMetrics() map[string]*Metrics.Metrics {
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

func (listener *TcpListener) CheckConnectionAttempts() uint64 {
	return listener.connectionAttempts.Load()
}
func (listener *TcpListener) GetConnectionAttempts() uint64 {
	return listener.connectionAttempts.Swap(0)
}

func (listener *TcpListener) CheckFailedConnections() uint64 {
	return listener.failedConnections.Load()
}
func (listener *TcpListener) GetFailedConnections() uint64 {
	return listener.failedConnections.Swap(0)
}

func (listener *TcpListener) CheckRejectedConnections() uint64 {
	return listener.rejectedConnections.Load()
}
func (listener *TcpListener) GetRejectedConnections() uint64 {
	return listener.rejectedConnections.Swap(0)
}

func (listener *TcpListener) CheckAcceptedConnections() uint64 {
	return listener.acceptedConnections.Load()
}
func (listener *TcpListener) GetAcceptedConnections() uint64 {
	return listener.acceptedConnections.Swap(0)
}
