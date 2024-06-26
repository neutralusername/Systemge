package Node

import "Systemge/Error"

// returns a map of custom command handlers for the command-line interface
func (node *Node) GetCustomCommandHandlers() map[string]func([]string) error {
	handlers := map[string]func([]string) error{
		"brokers":               node.handleBrokersCommand,
		"removeBroker":          node.handleRemoveBrokerCommand,
		"resolutions":           node.handleResolutionsCommand,
		"removeResolution":      node.handleRemoveTopicCommand,
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

func (node *Node) handleBrokersCommand(args []string) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	for _, brokerConnection := range node.activeBrokerConnections {
		println(brokerConnection.resolution.GetName() + " : " + brokerConnection.resolution.GetAddress())
	}
	return nil
}

func (node *Node) handleRemoveBrokerCommand(args []string) error {
	if len(args) != 1 {
		return Error.New("Invalid number of arguments", nil)
	}
	brokerAddress := args[0]
	err := node.RemoveBrokerConnection(brokerAddress)
	if err != nil {
		return Error.New("Error removing broker connection", err)
	}
	return nil
}

func (node *Node) handleResolutionsCommand(args []string) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	for topic, brokerConnection := range node.topicResolutions {
		println(topic + " : " + brokerConnection.resolution.GetName() + " : " + brokerConnection.resolution.GetAddress())
	}
	return nil
}

func (node *Node) handleRemoveTopicCommand(args []string) error {
	if len(args) != 1 {
		return Error.New("Invalid number of arguments", nil)
	}
	topic := args[0]
	err := node.RemoveTopicResolution(topic)
	if err != nil {
		return Error.New("Error removing topic resolution", err)
	}
	return nil
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
