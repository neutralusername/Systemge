package Broker

import "Systemge/Application"

// returns a map of custom command handlers for the command-line interface
func (server *Server) GetCustomCommandHandlers() map[string]Application.CustomCommandHandler {
	return map[string]Application.CustomCommandHandler{
		"brokerClients": func(args []string) error {
			server.mutex.Lock()
			for _, client := range server.connectedClients {
				println(client.name)
			}
			server.mutex.Unlock()
			return nil
		},
	}
}
