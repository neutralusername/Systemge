package Oauth2

import (
	"Systemge/Node"
)

func (server *App) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"oauth2Sessions": server.handleSessionsCommand,
	}
}

func (server *App) handleSessionsCommand(node *Node.Node, args []string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, session := range server.sessions {
		println(session.identity, session.sessionId)
	}
	return nil
}
