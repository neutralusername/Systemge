package WebsocketServer

func (server *WebsocketServer) WebsocketConnectionExists(websocketId string) bool {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	_, exists := server.websocketConnections[websocketId]
	return exists
}

func (server *WebsocketServer) GetWebsocketConnectionGroupCount(websocketId string) int {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	return len(server.websocketConnectionGroups[websocketId])
}

func (server *WebsocketServer) GetWebsocketConnectionCount() int {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	return len(server.websocketConnections)
}

func (server *WebsocketServer) GetWebsocketConnectionIds() []string {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	ids := make([]string, 0, len(server.websocketConnections))
	for id := range server.websocketConnections {
		ids = append(ids, id)
	}
	return ids
}
