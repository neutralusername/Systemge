package Node

// returns a map of command handlers for the command-line interface
func (node *Node) GetCommandHandlers() map[string]CommandHandler {
	handlers := map[string]CommandHandler{
		"websocketClients":             handleWebsocketClientsCommand,
		"websocketGroups":              handleWebsocketGroupsCommand,
		"WebsocketGroupClients":        handleWebsocketGroupClientsCommand,
		"websocketBlacklist":           handleWebsocketBlacklistCommand,
		"websocketWhitelist":           handleWebsocketWhitelistCommand,
		"addToWebsocketBlacklist":      handleAddToWebsocketBlacklistCommand,
		"addToWebsocketWhitelist":      handleAddToWebsocketWhitelistCommand,
		"removeFromWebsocketBlacklist": handleRemoveFromWebsocketBlacklistCommand,
		"removeFromWebsocketWhitelist": handleRemoveFromWebsocketWhitelistCommand,
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
	if len(args) < 1 {
		println("Usage: addToWebsocketBlacklist <ip>")
		return nil
	}
	node.addToWebsocketBlacklist(args...)
	return nil
}

func handleAddToWebsocketWhitelistCommand(node *Node, args []string) error {
	if len(args) < 1 {
		println("Usage: addToWebsocketWhitelist <ip>")
		return nil
	}
	node.addToWebsocketWhitelist(args...)
	return nil
}

func handleRemoveFromWebsocketBlacklistCommand(node *Node, args []string) error {
	if len(args) < 1 {
		println("Usage: removeFromWebsocketBlacklist <ip>")
		return nil
	}
	node.removeFromWebsocketBlacklist(args...)
	return nil
}

func handleRemoveFromWebsocketWhitelistCommand(node *Node, args []string) error {
	if len(args) < 1 {
		println("Usage: removeFromWebsocketWhitelist <ip>")
		return nil
	}
	node.removeFromWebsocketWhitelist(args...)
	return nil
}
