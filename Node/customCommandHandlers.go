package Node

import "Systemge/Module"

// returns a map of custom command handlers for the command-line interface
func (node *Node) GetCustomCommandHandlers() map[string]Module.CustomCommandHandler {
	handlers := map[string]Module.CustomCommandHandler{
		"websocketClients":      node.handleWebsocketClientsCommand,
		"websocketGroups":       node.handleWebsocketGroupsCommand,
		"WebsocketGroupClients": node.handleWebsocketGroupClientsCommand,
	}
	if node.application != nil {
		customHandlers := node.application.GetCustomCommandHandlers()
		for command, handler := range customHandlers {
			handlers[command] = func(args []string) error {
				return handler(node, args)
			}
		}
	}
	return handlers
}

func (node *Node) handleWebsocketClientsCommand(args []string) error {
	node.websocketMutex.Lock()
	for _, websocketClient := range node.websocketClients {
		println(websocketClient.GetId())
	}
	node.websocketMutex.Unlock()
	return nil
}

func (node *Node) handleWebsocketGroupsCommand(args []string) error {
	node.websocketMutex.Lock()
	for groupId := range node.WebsocketGroups {
		println(groupId)
	}
	node.websocketMutex.Unlock()
	return nil
}

func (node *Node) handleWebsocketGroupClientsCommand(args []string) error {
	if len(args) < 1 {
		println("Usage: groupClients <groupId>")
	}
	groupId := args[0]
	node.websocketMutex.Lock()
	group, ok := node.WebsocketGroups[groupId]
	node.websocketMutex.Unlock()
	if !ok {
		println("Group not found")
	} else {
		for _, websocketClient := range group {
			println(websocketClient.GetId())
		}
	}
	return nil
}
