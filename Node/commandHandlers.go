package Node

import "Systemge/Module"

// returns a map of custom command handlers for the command-line interface
func (node *Node) GetCommandHandlers() map[string]Module.CommandHandler {
	handlers := map[string]Module.CommandHandler{
		"websocketClients":      node.handleWebsocketClientsCommand,
		"websocketGroups":       node.handleWebsocketGroupsCommand,
		"WebsocketGroupClients": node.handleWebsocketGroupClientsCommand,
	}
	if cliComponent := node.GetCLIComponent(); cliComponent != nil {
		customHandlers := cliComponent.GetCommandHandlers()
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
	for groupId := range node.websocketGroups {
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
	group, ok := node.websocketGroups[groupId]
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
