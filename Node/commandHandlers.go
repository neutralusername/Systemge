package Node

// returns a map of command handlers for the command-line interface
func (node *Node) GetCommandHandlers() map[string]CommandHandler {
	handlers := map[string]CommandHandler{
		"start": func(node *Node, args []string) error {
			return node.Start()
		},
		"stop": func(node *Node, args []string) error {
			return node.Stop()
		},
	}
	if node.GetWebsocketComponent() != nil {
		handlers["websocketClients"] = handleWebsocketClientsCommand
		handlers["websocketGroups"] = handleWebsocketGroupsCommand
		handlers["websocketGroupClients"] = handleWebsocketGroupClientsCommand
		handlers["websocketBlacklist"] = handleWebsocketBlacklistCommand
		handlers["websocketWhitelist"] = handleWebsocketWhitelistCommand
		handlers["addWebsocketBlacklist"] = handleAddToWebsocketBlacklistCommand
		handlers["addWebsocketWhitelist"] = handleAddToWebsocketWhitelistCommand
		handlers["removeWebsocketBlacklist"] = handleRemoveFromWebsocketBlacklistCommand
		handlers["removeWebsocketWhitelist"] = handleRemoveFromWebsocketWhitelistCommand
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

func handleWebsocketBlacklistCommand(node *Node, args []string) error {
	node.websocketMutex.Lock()
	for ip := range node.websocketBlacklist {
		println(ip)
	}
	node.websocketMutex.Unlock()
	return nil
}

func handleWebsocketWhitelistCommand(node *Node, args []string) error {
	node.websocketMutex.Lock()
	for ip := range node.websocketWhitelist {
		println(ip)
	}
	node.websocketMutex.Unlock()
	return nil
}

func handleAddToWebsocketBlacklistCommand(node *Node, args []string) error {
	node.addToWebsocketBlacklist(args...)
	return nil
}

func handleAddToWebsocketWhitelistCommand(node *Node, args []string) error {
	node.addToWebsocketWhitelist(args...)
	return nil
}

func handleRemoveFromWebsocketBlacklistCommand(node *Node, args []string) error {
	node.removeFromWebsocketBlacklist(args...)
	return nil
}

func handleRemoveFromWebsocketWhitelistCommand(node *Node, args []string) error {
	node.removeFromWebsocketWhitelist(args...)
	return nil
}
