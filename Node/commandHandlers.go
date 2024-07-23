package Node

import "Systemge/Error"

// returns a map of command handlers for the command-line interface
func (node *Node) GetCommandHandlers() map[string]CommandHandler {
	handlers := map[string]CommandHandler{
		"start": func(node *Node, args []string) (string, error) {
			err := node.Start()
			if err != nil {
				return "", Error.New("Failed to start node \""+node.GetName()+"\": "+err.Error(), nil)
			}
			return "success", nil
		},
		"stop": func(node *Node, args []string) (string, error) {
			err := node.Stop()
			if err != nil {
				return "", Error.New("Failed to stop node \""+node.GetName()+"\": "+err.Error(), nil)
			}
			return "success", nil
		},
	}
	if node.websocket != nil {
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
	if node.http != nil {
		handlers["httpBlacklist"] = handleHttpBlacklistCommand
		handlers["httpWhitelist"] = handleHttpWhitelistCommand
		handlers["addHttpBlacklist"] = handleAddToHttpBlacklistCommand
		handlers["addHttpWhitelist"] = handleAddToHttpWhitelistCommand
		handlers["removeHttpBlacklist"] = handleRemoveFromHttpBlacklistCommand
		handlers["removeHttpWhitelist"] = handleRemoveFromHttpWhitelistCommand
	}
	if node.systemge != nil {
		handlers["topicResolutions"] = handleTopicResolutionsCommand
		handlers["brokerConnections"] = handleBrokerConnectionsCommand
	}
	if commandHandlerComponent := node.GetCommandHandlerComponent(); commandHandlerComponent != nil {
		commandHandlers := commandHandlerComponent.GetCommandHandlers()
		for command, commandHandler := range commandHandlers {
			handlers[command] = commandHandler
		}
	}
	return handlers
}

func handleTopicResolutionsCommand(node *Node, args []string) (string, error) {
	node.systemge.mutex.Lock()
	returnString := ""
	for topic, brokerConnection := range node.systemge.topicResolutions {
		returnString += topic + ":" + brokerConnection.endpoint.Address + ";"
	}
	node.systemge.mutex.Unlock()
	return returnString, nil
}

func handleBrokerConnectionsCommand(node *Node, args []string) (string, error) {
	node.systemge.mutex.Lock()
	returnString := ""
	for address := range node.systemge.brokerConnections {
		returnString += address + ";"
	}
	node.systemge.mutex.Unlock()
	return returnString, nil
}

func handleWebsocketClientsCommand(node *Node, args []string) (string, error) {
	node.websocket.mutex.Lock()
	defer node.websocket.mutex.Unlock()
	returnString := ""
	for _, websocketClient := range node.websocket.clients {
		returnString += websocketClient.GetId() + ";"
	}
	return returnString, nil
}

func handleWebsocketGroupsCommand(node *Node, args []string) (string, error) {
	node.websocket.mutex.Lock()
	returnString := ""
	for groupId := range node.websocket.groups {
		returnString += groupId + ";"
	}
	node.websocket.mutex.Unlock()
	return returnString, nil
}

func handleWebsocketGroupClientsCommand(node *Node, args []string) (string, error) {
	if len(args) < 1 {
		return "", Error.New("Invalid arguments", nil)
	}
	groupId := args[0]
	node.websocket.mutex.Lock()
	group, ok := node.websocket.groups[groupId]
	node.websocket.mutex.Unlock()
	if !ok {
		return "", Error.New("Group not found", nil)
	}
	returnString := ""
	for _, websocketClient := range group {
		returnString += websocketClient.GetId() + ";"
	}
	return returnString, nil
}

func handleWebsocketBlacklistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.websocket.httpServer.GetBlacklist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleWebsocketWhitelistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.websocket.httpServer.GetWhitelist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleAddToWebsocketBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleAddToWebsocketWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleRemoveFromWebsocketBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleRemoveFromWebsocketWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetWhitelist().Remove(ip)
	}
	return "success", nil
}

func handleHttpBlacklistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.http.server.GetBlacklist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleHttpWhitelistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.http.server.GetWhitelist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleAddToHttpBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleAddToHttpWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleRemoveFromHttpBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleRemoveFromHttpWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetWhitelist().Remove(ip)
	}
	return "success", nil
}
