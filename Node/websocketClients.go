package Node

import (
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (websocket *websocketComponent) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient {
	websocket.mutex.Lock()
	defer websocket.mutex.Unlock()
	websocketId := Tools.RandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := websocket.clients[websocketId]; exists; {
		websocketId = Tools.RandomString(16, Tools.ALPHA_NUMERIC)
	}
	websocketClient := websocket.newWebsocketClient(websocketId, websocketConn)
	websocket.clients[websocketId] = websocketClient
	websocket.clientGroups[websocketId] = make(map[string]bool)
	return websocketClient
}

func (websocket *websocketComponent) removeWebsocketClient(websocketClient *WebsocketClient) {
	websocket.mutex.Lock()
	defer websocket.mutex.Unlock()
	delete(websocket.clients, websocketClient.GetId())
	for groupId := range websocket.clientGroups[websocketClient.GetId()] {
		delete(websocket.clientGroups[websocketClient.GetId()], groupId)
		delete(websocket.groups[groupId], websocketClient.GetId())
		if len(websocket.groups[groupId]) == 0 {
			delete(websocket.groups, groupId)
		}
	}
}

// WebsocketClientExists returns true if a websocket client with the given id exists.
func (node *Node) WebsocketClientExists(websocketId string) bool {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.RLock()
		defer node.websocket.mutex.RUnlock()
		_, exists := node.websocket.clients[websocketId]
		return exists
	}
	return false
}

// GetWebsocketClientGroupCount returns the number of groups a websocket client is in (0 if the client does not exist).
func (node *Node) GetWebsocketClientGroupCount(websocketId string) int {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.RLock()
		defer node.websocket.mutex.RUnlock()
		return len(node.websocket.clientGroups[websocketId])
	}
	return 0
}

// GetWebsocketClientCount returns the number of connected websocket clients.
func (node *Node) GetWebsocketClientCount() int {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.RLock()
		defer node.websocket.mutex.RUnlock()
		return len(node.websocket.clients)
	}
	return 0
}

// GetWebsocketGroupCount returns the number of websocket groups.
func (node *Node) GetWebsocketGroupCount() int {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.RLock()
		defer node.websocket.mutex.RUnlock()
		return len(node.websocket.groups)
	}
	return 0
}
