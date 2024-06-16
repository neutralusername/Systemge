package WebsocketServer

import (
	"Systemge/Randomizer"
	"Systemge/WebsocketClient"

	"github.com/gorilla/websocket"
)

func (server *Server) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient.Client {
	server.acquireMutex()
	defer server.releaseMutex()
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
	server.acquireMutex()
	defer server.releaseMutex()
	delete(server.clients, client.GetId())
	for groupId := range server.clientGroups[client.GetId()] {
		delete(server.clientGroups[client.GetId()], groupId)
		delete(server.groups[groupId], client.GetId())
		if len(server.groups[groupId]) == 0 {
			delete(server.groups, groupId)
		}
	}
}

func (server *Server) ClientExists(websocketId string) bool {
	server.acquireMutex()
	defer server.releaseMutex()
	_, exists := server.clients[websocketId]
	return exists
}
