package HTTPServer

func (server *HTTPServer) GetMetrics_() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"httpServer": {
			"requestCounter": server.GetHTTPRequestCounter(),
		},
	}
}
func (server *HTTPServer) GetMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"httpServer": {
			"requestCounter": server.RetrieveHTTPRequestCounter(),
		},
	}
}

func (server *HTTPServer) RetrieveHTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *HTTPServer) GetHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}
