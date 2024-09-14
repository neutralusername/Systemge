package DashboardServer

import (
	"runtime"
	"time"

	"github.com/neutralusername/Systemge/DashboardHelpers"
)

func (server *Server) getHeapUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.HeapSys
}

func (server *Server) getGoroutineCount() int {
	return runtime.NumGoroutine()
}

func (server *Server) retrieveDashboardClientMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		DashboardHelpers.DASHBOARD_METRICSTYPE_SYSTEMGE:      server.RetrieveSystemgeMetrics(),
		DashboardHelpers.DASHBOARD_METRICSTYPE_WEBSOCKET:     server.RetrieveWebsocketMetrics(),
		DashboardHelpers.DASHBOARD_METRICSTYPE_HTTP:          server.RetrieveHttpMetrics(),
		DashboardHelpers.DASHBOARD_METRICSTYPE_RESOURCEUSAGE: server.GetResourceUsageMetrics(),
	}
}

func (server *Server) addMetricsToDashboardClient(metrics map[string]map[string]uint64) {
	if server.dashboardClient.Metrics == nil {
		server.dashboardClient.Metrics = map[string]map[string][]*DashboardHelpers.MetricsEntry{}
	}
	for metricsType, metricsKeyValuePairs := range metrics {
		if server.dashboardClient.Metrics[metricsType] == nil {
			server.dashboardClient.Metrics[metricsType] = map[string][]*DashboardHelpers.MetricsEntry{}
		}
		for key, value := range metricsKeyValuePairs {
			server.dashboardClient.Metrics[metricsType][key] = append(server.dashboardClient.Metrics[metricsType][key], &DashboardHelpers.MetricsEntry{
				Value: value,
				Time:  time.Now(),
			})
			if len(server.dashboardClient.Metrics[metricsType][key]) > server.config.MaxMetricsCacheValues {
				server.dashboardClient.Metrics[metricsType][key] = server.dashboardClient.Metrics[metricsType][key][1:]
			}
		}
	}
}

func (server *Server) GetResourceUsageMetrics() map[string]uint64 {
	return map[string]uint64{
		"heap_usage":      server.getHeapUsage(),
		"goroutine_count": uint64(server.getGoroutineCount()),
	}
}

func (server *Server) GetSystemgeMetrics() map[string]uint64 {
	return server.systemgeServer.GetMetrics()
}
func (server *Server) RetrieveSystemgeMetrics() map[string]uint64 {
	return server.systemgeServer.RetrieveMetrics()
}

func (server *Server) GetWebsocketMetrics() map[string]uint64 {
	return server.websocketServer.GetMetrics()
}
func (server *Server) RetrieveWebsocketMetrics() map[string]uint64 {
	return server.websocketServer.RetrieveMetrics()
}

func (server *Server) GetHttpMetrics() map[string]uint64 {
	return server.httpServer.GetMetrics()
}
func (server *Server) RetrieveHttpMetrics() map[string]uint64 {
	return server.httpServer.RetrieveMetrics()
}
