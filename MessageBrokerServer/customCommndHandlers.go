package MessageBrokerServer

// returns a map of custom command handlers for the command-line interface
func (server *Server) GetCustomCommandHandlers() map[string]func([]string) error {
	return map[string]func([]string) error{
		"clientCount": func(args []string) error {
			clientsCount := server.GetClientCount()
			println(clientsCount)
			return nil
		},
	}
}
