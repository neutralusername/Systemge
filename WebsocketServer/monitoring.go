package WebsocketServer

import "Systemge/WebsocketClient"

func (server *Server) GetOnlineWebsocketClientsCount() int {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	return len(server.websocketClients)
}

func (server *Server) GetOnlineWebsocketIds() []string {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	ids := make([]string, 0)
	for id := range server.websocketClients {
		ids = append(ids, id)
	}
	return ids
}

func (server *Server) GetWebsocketClient(id string) *WebsocketClient.Client {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	return server.websocketClients[id]
}
