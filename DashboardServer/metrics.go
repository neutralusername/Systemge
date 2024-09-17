package DashboardServer

import (
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *Server) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("resource_usage", Metrics.New(
		map[string]uint64{
			"heap_usage":      Helpers.HeapUsage(),
			"goroutine_count": uint64(Helpers.GoroutineCount()),
		},
	))
	metricsTypes.Merge(server.websocketServer.GetMetrics())
	metricsTypes.Merge(server.httpServer.GetMetrics())
	metricsTypes.Merge(server.systemgeServer.GetMetrics())
	return metricsTypes
}

func (server *Server) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("resource_usage", Metrics.New(
		map[string]uint64{
			"heap_usage":      Helpers.HeapUsage(),
			"goroutine_count": uint64(Helpers.GoroutineCount()),
		},
	))
	metricsTypes.Merge(server.websocketServer.CheckMetrics())
	metricsTypes.Merge(server.httpServer.CheckMetrics())
	metricsTypes.Merge(server.systemgeServer.CheckMetrics())
	return metricsTypes
}
