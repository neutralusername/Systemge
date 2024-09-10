package DashboardServer

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
	return map[string]uint64{
		"http_request_count": server.httpServer.GetHTTPRequestCounter(),
	}
}
func (server *Server) RetrieveHttpMetrics() map[string]uint64 {
	return map[string]uint64{
		"http_request_count": server.httpServer.RetrieveHTTPRequestCounter(),
	}
}
