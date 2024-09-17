package HTTPServer

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *HTTPServer) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("http_server", Metrics.New(
		map[string]uint64{
			"request_counter": server.CheckHTTPRequestCounter(),
		},
	))
	return metricsTypes
}
func (server *HTTPServer) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("http_server", Metrics.New(
		map[string]uint64{
			"request_counter": server.GetTTPRequestCounter(),
		},
	))
	return metricsTypes
}

func (server *HTTPServer) GetTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *HTTPServer) CheckHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}
