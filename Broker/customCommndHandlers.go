package Broker

import "Systemge/Application"

// returns a map of custom command handlers for the command-line interface
func (server *Server) GetCustomCommandHandlers() map[string]Application.CustomCommandHandler {
	return map[string]Application.CustomCommandHandler{
		"brokerClients": func(args []string) error {
			server.operationMutex.Lock()
			for _, clientConnection := range server.clientConnections {
				println(clientConnection.name)
			}
			server.operationMutex.Unlock()
			return nil
		},
	}
}
