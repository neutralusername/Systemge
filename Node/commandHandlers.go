package Node

import (
	"github.com/neutralusername/Systemge/Error"
)

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
		handlers["websocketAddBlacklist"] = handleWebsocketAddBlacklistCommand
		handlers["websocketAddWhitelist"] = handleWebsocketAddWhitelistCommand
		handlers["websocketRemoveBlacklist"] = handleWebsocketRemoveBlacklistCommand
		handlers["websocketRemoveWhitelist"] = handleWebsocketRemoveWhitelistCommand
	}
	if node.http != nil {
		handlers["httpBlacklist"] = handleHttpBlacklistCommand
		handlers["httpWhitelist"] = handleHttpWhitelistCommand
		handlers["httpAddBlacklist"] = handleHttpAddBlacklistCommand
		handlers["httpAddWhitelist"] = handleHttpAddWhitelistCommand
		handlers["HttpRemoveBlacklist"] = handleHttpRemoveBlacklistCommand
		handlers["HttpRemoveWhitelist"] = handleHttpRemoveWhitelistCommand
	}
	if node.systemgeServer != nil {

	}

	if commandHandlerComponent := node.GetCommandHandlerComponent(); commandHandlerComponent != nil {
		commandHandlers := commandHandlerComponent.GetCommandHandlers()
		for command, commandHandler := range commandHandlers {
			handlers[command] = commandHandler
		}
	}
	return handlers
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

func handleWebsocketAddBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleWebsocketAddWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleWebsocketRemoveBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.websocket.httpServer.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleWebsocketRemoveWhitelistCommand(node *Node, args []string) (string, error) {
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

func handleHttpAddBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleHttpAddWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleHttpRemoveBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleHttpRemoveWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.http.server.GetWhitelist().Remove(ip)
	}
	return "success", nil
}
