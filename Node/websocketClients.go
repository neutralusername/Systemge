package Node

import (
	"Systemge/Tools"

	"github.com/gorilla/websocket"
)

func (websocket *websocketComponent) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient {
	websocket.mutex.Lock()
	defer websocket.mutex.Unlock()
	websocketId := "#" + Tools.RandomString(16, Tools.ALPHA_NUMERIC)
	for _, exists := websocket.clients[websocketId]; exists; {
		websocketId = "#" + Tools.RandomString(16, Tools.ALPHA_NUMERIC)
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

func (node *Node) WebsocketClientExists(websocketId string) bool {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.Lock()
		defer node.websocket.mutex.Unlock()
		_, exists := node.websocket.clients[websocketId]
		return exists
	}
	return false
}
