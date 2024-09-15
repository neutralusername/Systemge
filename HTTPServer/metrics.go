package HTTPServer

func (server *HTTPServer) CheckMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"http_server": {
			"request_counter": server.CheckHTTPRequestCounter(),
		},
	}
}
func (server *HTTPServer) GetMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"http_server": {
			"request_counter": server.GetTTPRequestCounter(),
		},
	}
}

func (server *HTTPServer) GetTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *HTTPServer) CheckHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}
