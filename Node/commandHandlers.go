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
	if node.GetHTTPComponent() != nil {
		handlers["httpBlacklist"] = handleHttpBlacklistCommand
		handlers["httpWhitelist"] = handleHttpWhitelistCommand
		handlers["addHttpBlacklist"] = handleAddToHttpBlacklistCommand
		handlers["addHttpWhitelist"] = handleAddToHttpWhitelistCommand
		handlers["removeHttpBlacklist"] = handleRemoveFromHttpBlacklistCommand
		handlers["removeHttpWhitelist"] = handleRemoveFromHttpWhitelistCommand
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

func handleHttpBlacklistCommand(node *Node, args []string) error {
	node.httpMutex.Lock()
	for ip := range node.httpBlacklist {
		println(ip)
	}
	node.httpMutex.Unlock()
	return nil
}

func handleHttpWhitelistCommand(node *Node, args []string) error {
	node.httpMutex.Lock()
	for ip := range node.httpWhitelist {
		println(ip)
	}
	node.httpMutex.Unlock()
	return nil
}

func handleAddToHttpBlacklistCommand(node *Node, args []string) error {
	node.addToHttpBlacklist(args...)
	return nil
}

func handleAddToHttpWhitelistCommand(node *Node, args []string) error {
	node.addToHttpWhitelist(args...)
	return nil
}

func handleRemoveFromHttpBlacklistCommand(node *Node, args []string) error {
	node.removeFromHttpBlacklist(args...)
	return nil
}

func handleRemoveFromHttpWhitelistCommand(node *Node, args []string) error {
	node.removeFromHttpWhitelist(args...)
	return nil
}
