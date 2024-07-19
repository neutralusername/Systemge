package Oauth2

import (
	"Systemge/Node"
)

func (server *Server) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"oauth2Sessions":  server.handleSessionsCommand,
		"blacklist":       server.handleBlacklistCommand,
		"whitelist":       server.handleWhitelistCommand,
		"addWhitelist":    server.handleAddWhitelistCommand,
		"addBlacklist":    server.handleAddBlacklistCommand,
		"removeWhitelist": server.handleRemoveWhitelistCommand,
		"removeBlacklist": server.handleRemoveBlacklistCommand,
	}
}

func (server *Server) handleSessionsCommand(node *Node.Node, args []string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, session := range server.sessions {
		println(session.identity, session.sessionId)
	}
	return nil
}

func (server *Server) handleBlacklistCommand(node *Node.Node, args []string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for ip := range server.blacklist {
		println(ip)
	}
	return nil
}

func (server *Server) handleWhitelistCommand(node *Node.Node, args []string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for ip := range server.whitelist {
		println(ip)
	}
	return nil
}

func (server *Server) handleAddWhitelistCommand(node *Node.Node, args []string) error {
	server.addToWhitelist(args...)
	return nil
}

func (server *Server) handleAddBlacklistCommand(node *Node.Node, args []string) error {
	server.addToBlacklist(args...)
	return nil
}

func (server *Server) handleRemoveWhitelistCommand(node *Node.Node, args []string) error {
	server.removeFromWhitelist(args...)
	return nil
}

func (server *Server) handleRemoveBlacklistCommand(node *Node.Node, args []string) error {
	server.removeFromBlacklist(args...)
	return nil
}
