package HTTPServer

func (server *HTTPServer) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"requestCounter": server.GetHTTPRequestCounter(),
	}
}
func (server *HTTPServer) RetrieveMetrics() map[string]uint64 {
	return map[string]uint64{
		"requestCounter": server.RetrieveHTTPRequestCounter(),
	}
}

func (server *HTTPServer) RetrieveHTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *HTTPServer) GetHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}
