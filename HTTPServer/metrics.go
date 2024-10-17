package HTTPServer

import "github.com/neutralusername/Systemge/Tools"

func (server *HTTPServer) CheckMetrics() Tools.MetricsTypes {
	metricsTypes := Tools.NewMetricsTypes()
	metricsTypes.AddMetrics("http_server", Tools.NewMetrics(
		map[string]uint64{
			"request_counter": server.RequestCounter.Load(),
		},
	))
	return metricsTypes
}
func (server *HTTPServer) GetMetrics() Tools.MetricsTypes {
	metricsTypes := Tools.NewMetricsTypes()
	metricsTypes.AddMetrics("http_server", Tools.NewMetrics(
		map[string]uint64{
			"request_counter": server.RequestCounter.Swap(0),
		},
	))
	return metricsTypes
}
