package Client

import (
	"Systemge/Randomizer"
	"Systemge/WebsocketClient"

	"github.com/gorilla/websocket"
)

func (client *Client) addWebsocketConn(websocketConn *websocket.Conn) *WebsocketClient.Client {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	websocketId := "#" + client.randomizer.GenerateRandomString(16, Randomizer.ALPHA_NUMERIC)
	for _, exists := client.websocketClients[websocketId]; exists; {
		websocketId = "#" + client.randomizer.GenerateRandomString(16, Randomizer.ALPHA_NUMERIC)
	}
	websocketClient := WebsocketClient.New(websocketId, websocketConn, func(websocketClient *WebsocketClient.Client) {
		client.websocketApplication.OnDisconnectHandler(client, websocketClient)
		client.removeWebsocketClient(websocketClient)
	})
	client.websocketClients[websocketId] = websocketClient
	client.websocketClientGroups[websocketId] = make(map[string]bool)
	return websocketClient
}

func (client *Client) removeWebsocketClient(websocketClient *WebsocketClient.Client) {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	delete(client.websocketClients, websocketClient.GetId())
	for groupId := range client.websocketClientGroups[websocketClient.GetId()] {
		delete(client.websocketClientGroups[websocketClient.GetId()], groupId)
		delete(client.groups[groupId], websocketClient.GetId())
		if len(client.groups[groupId]) == 0 {
			delete(client.groups, groupId)
		}
	}
}

func (client *Client) ClientExists(websocketId string) bool {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	_, exists := client.websocketClients[websocketId]
	return exists
}
