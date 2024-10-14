package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/WebsocketConnection"
)

func (server *WebsocketServer[O]) GetWebsocketClient(sessionId string) (*WebsocketConnection.WebsocketConnection, error) {
	session := server.sessionManager.GetSession(sessionId)
	if session == nil {
		return nil, errors.New("Session not found")
	}
	websocketClient, ok := session.Get("websocketClient")
	if !ok {
		return nil, errors.New("WebsocketClient not found")
	}
	return websocketClient.(*WebsocketConnection.WebsocketConnection), nil
}

func (server *WebsocketServer[O]) GetIdentityWebsocketClients(identity string) []*WebsocketConnection.WebsocketConnection {
	clients := []*WebsocketConnection.WebsocketConnection{}
	sessions := server.sessionManager.GetIdentitySessions(identity)
	for _, session := range sessions {
		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			continue
		}
		clients = append(clients, websocketClient.(*WebsocketConnection.WebsocketConnection))
	}
	return clients
}

func (server *WebsocketServer[O]) GetWebsocketClients() []*WebsocketConnection.WebsocketConnection {
	clients := []*WebsocketConnection.WebsocketConnection{}
	sessions := server.sessionManager.GetSessions()
	for _, session := range sessions {
		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			continue
		}
		clients = append(clients, websocketClient.(*WebsocketConnection.WebsocketConnection))
	}
	return clients
}
