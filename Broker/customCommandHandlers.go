package Broker

// returns a map of custom command handlers for the command-line interface
func (server *Server) GetCustomCommandHandlers() map[string]func([]string) error {
	return map[string]func([]string) error{
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
