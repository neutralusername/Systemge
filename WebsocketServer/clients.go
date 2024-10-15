package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/ConnectionWebsocket"
)

func (server *WebsocketServer[O]) GetWebsocketClient(sessionId string) (*ConnectionWebsocket.WebsocketConnection, error) {
	session := server.sessionManager.GetSession(sessionId)
	if session == nil {
		return nil, errors.New("Session not found")
	}
	websocketClient, ok := session.Get("websocketClient")
	if !ok {
		return nil, errors.New("WebsocketClient not found")
	}
	return websocketClient.(*ConnectionWebsocket.WebsocketConnection), nil
}

func (server *WebsocketServer[O]) GetIdentityWebsocketClients(identity string) []*ConnectionWebsocket.WebsocketConnection {
	clients := []*ConnectionWebsocket.WebsocketConnection{}
	sessions := server.sessionManager.GetIdentitySessions(identity)
	for _, session := range sessions {
		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			continue
		}
		clients = append(clients, websocketClient.(*ConnectionWebsocket.WebsocketConnection))
	}
	return clients
}

func (server *WebsocketServer[O]) GetWebsocketClients() []*ConnectionWebsocket.WebsocketConnection {
	clients := []*ConnectionWebsocket.WebsocketConnection{}
	sessions := server.sessionManager.GetSessions()
	for _, session := range sessions {
		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			continue
		}
		clients = append(clients, websocketClient.(*ConnectionWebsocket.WebsocketConnection))
	}
	return clients
}
