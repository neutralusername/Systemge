package WebsocketServer

import "Systemge/Application"

// returns a map of custom command handlers for the command-line interface
func (server *Server) GetCustomCommandHandlers() map[string]Application.CustomCommandHandler {
	return map[string]Application.CustomCommandHandler{
		"websocketClients": func(args []string) error {
			server.acquireMutex()
			for _, client := range server.clients {
				println(client.GetId())
			}
			server.releaseMutex()
			return nil
		},
		"groups": func(args []string) error {
			server.acquireMutex()
			for groupId := range server.groups {
				println(groupId)
			}
			server.releaseMutex()
			return nil
		},
		"groupClients": func(args []string) error {
			if len(args) < 1 {
				println("Usage: groupClients <groupId>")
			}
			groupId := args[0]
			server.acquireMutex()
			group, ok := server.groups[groupId]
			if !ok {
				println("Group not found")
			} else {
				for _, client := range group {
					println(client.GetId())
				}
			}
			return nil
		},
	}
}
