package WebsocketServer

import (
	"Systemge/Randomizer"
	"Systemge/WebsocketClient"

	"github.com/gorilla/websocket"
)

func (server *Server) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient.Client {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	websocketId := "#" + server.randomizer.GenerateRandomString(16, Randomizer.ALPHA_NUMERIC)
	for _, exists := server.clients[websocketId]; exists; {
		websocketId = "#" + server.randomizer.GenerateRandomString(16, Randomizer.ALPHA_NUMERIC)
	}
	client := WebsocketClient.New(websocketId, websocketConn, func(client *WebsocketClient.Client) {
		server.websocketApplication.OnDisconnectHandler(client)
		server.removeClient(client)
	})
	server.clients[websocketId] = client
	server.clientGroups[websocketId] = make(map[string]bool)
	return client
}

func (server *Server) removeClient(client *WebsocketClient.Client) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	delete(server.clients, client.GetId())
	for groupId := range server.clientGroups[client.GetId()] {
		server.RemoveFromGroup(groupId, client.GetId())
	}
}

func (server *Server) ClientExists(websocketId string) bool {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	_, exists := server.clients[websocketId]
	return exists
}
