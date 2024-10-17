package httpServer

import "github.com/neutralusername/systemge/tools"

func (server *HTTPServer) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("http_server", tools.NewMetrics(
		map[string]uint64{
			"request_counter": server.RequestCounter.Load(),
		},
	))
	return metricsTypes
}
func (server *HTTPServer) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("http_server", tools.NewMetrics(
		map[string]uint64{
			"request_counter": server.RequestCounter.Swap(0),
		},
	))
	return metricsTypes
}
