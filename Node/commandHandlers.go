package Node

// returns a map of command handlers for the command-line interface
func (node *Node) GetCommandHandlers() map[string]CommandHandler {
	handlers := map[string]CommandHandler{
		"websocketClients":      handleWebsocketClientsCommand,
		"websocketGroups":       handleWebsocketGroupsCommand,
		"WebsocketGroupClients": handleWebsocketGroupClientsCommand,
	}
	if commandHandlerComponent := node.GetCommandHandlerComponent(); commandHandlerComponent != nil {
		commandHandlers := commandHandlerComponent.GetCommandHandlers()
		for command, commandHandler := range commandHandlers {
			handlers[command] = commandHandler
		}
	}
	return handlers
}

func handleWebsocketClientsCommand(node *Node, args []string) error {
	node.websocketMutex.Lock()
	for _, websocketClient := range node.websocketClients {
		println(websocketClient.GetId())
	}
	node.websocketMutex.Unlock()
	return nil
}

func handleWebsocketGroupsCommand(node *Node, args []string) error {
	node.websocketMutex.Lock()
	for groupId := range node.websocketGroups {
		println(groupId)
	}
	node.websocketMutex.Unlock()
	return nil
}

func handleWebsocketGroupClientsCommand(node *Node, args []string) error {
	if len(args) < 1 {
		println("Usage: groupClients <groupId>")
	}
	groupId := args[0]
	node.websocketMutex.Lock()
	group, ok := node.websocketGroups[groupId]
	node.websocketMutex.Unlock()
	if !ok {
		println("Group not found")
		return nil
	}
	for _, websocketClient := range group {
		println(websocketClient.GetId())
	}
	return nil
}
