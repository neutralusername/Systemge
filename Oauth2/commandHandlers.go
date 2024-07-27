package Oauth2

import (
	"github.com/neutralusername/Systemge/Node"
)

func (server *App) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"oauth2Sessions": server.handleSessionsCommand,
	}
}

func (server *App) handleSessionsCommand(node *Node.Node, args []string) (string, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	returnString := ""
	for _, session := range server.sessions {
		returnString += session.identity + ":" + session.sessionId + ";"
	}
	return returnString, nil
}
