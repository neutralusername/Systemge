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
	for metricsType, metricsKeyValuePairs := range metrics {
		if server.dashboardClient.Metrics[metricsType] == nil {
			server.dashboardClient.Metrics[metricsType] = map[string][]*DashboardHelpers.MetricsEntry{}
		}
		for key, value := range metricsKeyValuePairs {
			server.dashboardClient.Metrics[metricsType][key] = append(server.dashboardClient.Metrics[metricsType][key], &DashboardHelpers.MetricsEntry{
				Value: value,
				Time:  time.Now(),
			})
			if len(server.dashboardClient.Metrics[metricsType][key]) > server.config.MaxMetricEntries {
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
	metrics := server.systemgeServer.GetMetrics_()
	for _, value := range metrics {
		return value
	}
	return nil
}
func (server *Server) RetrieveSystemgeMetrics() map[string]uint64 {
	metrics := server.systemgeServer.GetMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}

func (server *Server) GetWebsocketMetrics() map[string]uint64 {
	metrics := server.websocketServer.GetMetrics_()
	for _, value := range metrics {
		return value
	}
	return nil
}
func (server *Server) RetrieveWebsocketMetrics() map[string]uint64 {
	metrics := server.websocketServer.GetMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}

func (server *Server) GetHttpMetrics() map[string]uint64 {
	metrics := server.httpServer.GetMetrics_()
	for _, value := range metrics {
		return value
	}
	return nil
}
func (server *Server) RetrieveHttpMetrics() map[string]uint64 {
	metrics := server.httpServer.GetMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}
