package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (server *Server) addWebsocketConn(websocketConn *websocket.Conn) *Client {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	websocketId := Tools.RandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := server.clients[websocketId]; exists; {
		websocketId = Tools.RandomString(16, Tools.ALPHA_NUMERIC)
	}
	websocketClient := server.newWebsocketClient(websocketId, websocketConn)
	server.clients[websocketId] = websocketClient
	server.clientGroups[websocketId] = make(map[string]bool)
	return websocketClient
}

func (server *Server) removeWebsocketClient(websocketClient *Client) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.clients, websocketClient.GetId())
	for groupId := range server.clientGroups[websocketClient.GetId()] {
		delete(server.clientGroups[websocketClient.GetId()], groupId)
		delete(server.groups[groupId], websocketClient.GetId())
		if len(server.groups[groupId]) == 0 {
			delete(server.groups, groupId)
		}
	}
}

// WebsocketClientExists returns true if a websocket client with the given id exists.
func (server *Server) WebsocketClientExists(websocketId string) bool {
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	_, exists := server.clients[websocketId]
	return exists
}

// GetWebsocketClientGroupCount returns the number of groups a websocket client is in (0 if the client does not exist).
func (server *Server) GetWebsocketClientGroupCount(websocketId string) int {
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	return len(server.clientGroups[websocketId])
}

// GetWebsocketClientCount returns the number of connected websocket clients.
func (websocket *Server) GetWebsocketClientCount() int {
	websocket.mutex.RLock()
	defer websocket.mutex.RUnlock()
	return len(websocket.clients)
}

// GetWebsocketGroupCount returns the number of websocket groups.
func (websocket *Server) GetWebsocketGroupCount() int {
	websocket.mutex.RLock()
	defer websocket.mutex.RUnlock()
	return len(websocket.groups)
}
