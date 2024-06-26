package Broker

// returns a map of custom command handlers for the command-line interface
func (server *Server) GetCustomCommandHandlers() map[string]func([]string) error {
	return map[string]func([]string) error{
		"brokerNodes": func(args []string) error {
			server.operationMutex.Lock()
			for _, nodeConnection := range server.nodeConnections {
				println(nodeConnection.name)
			}
			server.operationMutex.Unlock()
			return nil
		},
	}
}
