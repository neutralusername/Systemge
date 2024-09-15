package HTTPServer

import (
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

func (server *HTTPServer) CheckMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"http_server": {
			KeyValuePairs: map[string]uint64{
				"request_counter": server.CheckHTTPRequestCounter(),
			},
			Time: time.Now(),
		},
	}
}
func (server *HTTPServer) GetMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"http_server": {
			KeyValuePairs: map[string]uint64{
				"request_counter": server.GetTTPRequestCounter(),
			},
			Time: time.Now(),
		},
	}
}

func (server *HTTPServer) GetTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *HTTPServer) CheckHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}
