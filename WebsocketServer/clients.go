package WebsocketServer

// ClientExists returns true if a client with the given id exists.
func (server *WebsocketServer) ClientExists(websocketId string) bool {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	_, exists := server.clients[websocketId]
	return exists
}

// GetClientGroupCount returns the number of groups a client is in (0 if the client does not exist).
func (server *WebsocketServer) GetClientGroupCount(websocketId string) int {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	return len(server.clientGroups[websocketId])
}

// GetClientCount returns the number of connected clients.
func (server *WebsocketServer) GetClientCount() int {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	return len(server.clients)
}

func (server *WebsocketServer) removeClient(client *WebsocketClient) {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	delete(server.clients, client.GetId())
	for groupId := range server.clientGroups[client.GetId()] {
		delete(server.clientGroups[client.GetId()], groupId)
		delete(server.groups[groupId], client.GetId())
		if len(server.groups[groupId]) == 0 {
			delete(server.groups, groupId)
		}
	}
}
