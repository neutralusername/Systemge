package WebsocketServer

// ClientExists returns true if a client with the given id exists.
func (server *WebsocketServer) ClientExists(websocketId string) bool {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	_, exists := server.websocketConnections[websocketId]
	return exists
}

// GetClientGroupCount returns the number of groups a client is in (0 if the client does not exist).
func (server *WebsocketServer) GetClientGroupCount(websocketId string) int {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	return len(server.websocketConnectionGroups[websocketId])
}

// GetClientCount returns the number of connected clients.
func (server *WebsocketServer) GetClientCount() int {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	return len(server.websocketConnections)
}

func (server *WebsocketServer) removeWebsocketConnection(client *WebsocketConnection) {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()
	delete(server.websocketConnections, client.GetId())
	for groupId := range server.websocketConnectionGroups[client.GetId()] {
		delete(server.websocketConnectionGroups[client.GetId()], groupId)
		delete(server.groups[groupId], client.GetId())
		if len(server.groups[groupId]) == 0 {
			delete(server.groups, groupId)
		}
	}
}
