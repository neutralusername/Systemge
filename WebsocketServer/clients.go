package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) GetWebsocketClient(sessionId string) (*WebsocketClient.WebsocketClient, error) {
	session := server.sessionManager.GetSession(sessionId)
	if session == nil {
		return nil, errors.New("Session not found")
	}
	websocketClient, ok := session.Get("websocketClient")
	if !ok {
		return nil, errors.New("WebsocketClient not found")
	}
	return websocketClient.(*WebsocketClient.WebsocketClient), nil
}

func (server *WebsocketServer) GetWebsocketClients() []*WebsocketClient.WebsocketClient {
	clients := []*WebsocketClient.WebsocketClient{}
	sessions := server.sessionManager.GetSessions("")
	for _, session := range sessions {
		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			continue
		}
		clients = append(clients, websocketClient.(*WebsocketClient.WebsocketClient))
	}
	return clients
}
