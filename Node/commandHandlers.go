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
	for ip := range node.websocketBlacklist.GetElements() {
		println(ip)
	}
	return nil
}

func handleWebsocketWhitelistCommand(node *Node, args []string) error {
	for ip := range node.websocketWhitelist.GetElements() {
		println(ip)
	}
	return nil
}

func handleAddToWebsocketBlacklistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.websocketBlacklist.Add(ip)
	}
	return nil
}

func handleAddToWebsocketWhitelistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.websocketWhitelist.Add(ip)
	}
	return nil
}

func handleRemoveFromWebsocketBlacklistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.websocketBlacklist.Remove(ip)
	}
	return nil
}

func handleRemoveFromWebsocketWhitelistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.websocketWhitelist.Remove(ip)
	}
	return nil
}

func handleHttpBlacklistCommand(node *Node, args []string) error {
	for ip := range node.httpBlacklist.GetElements() {
		println(ip)
	}
	return nil
}

func handleHttpWhitelistCommand(node *Node, args []string) error {
	for ip := range node.httpWhitelist.GetElements() {
		println(ip)
	}
	return nil
}

func handleAddToHttpBlacklistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.httpBlacklist.Add(ip)
	}
	return nil
}

func handleAddToHttpWhitelistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.httpWhitelist.Add(ip)
	}
	return nil
}

func handleRemoveFromHttpBlacklistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.httpBlacklist.Remove(ip)
	}
	return nil
}

func handleRemoveFromHttpWhitelistCommand(node *Node, args []string) error {
	for _, ip := range args {
		node.httpWhitelist.Remove(ip)
	}
	return nil
}
