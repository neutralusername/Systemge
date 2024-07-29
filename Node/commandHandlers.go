package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
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
	if node.systemge != nil {
		handlers["systemgeTopicResolutions"] = handleSystemgeTopicResolutionsCommand
		handlers["systemgeBrokerConnection"] = handleSystemgeBrokerConnectionsCommand
	}
	if node.broker != nil {
		handlers["brokerNodeConnections"] = handleBrokerNodeConnectionsCommand
		handlers["brokerNodeConnectionCount"] = handleBrokerNodeConnectionCountCommand
		handlers["brokerSyncTopics"] = handleBrokerSyncTopicsCommand
		handlers["brokerAsyncTopics"] = handleBrokerAsyncTopicsCommand
		handlers["brokerWhitelist"] = handleBrokerWhitelistCommand
		handlers["brokerBlacklist"] = handleBrokerBlacklistCommand
		handlers["brokerConfigWhitelist"] = handleBrokerConfigWhitelistCommand
		handlers["brokerConfigBlacklist"] = handleBrokerConfigBlacklistCommand
		handlers["brokerAddSyncTopic"] = handleBrokerAddSyncTopicCommand
		handlers["brokerAddAsyncTopic"] = handleBrokerAddAsyncTopicCommand
		handlers["brokerRemoveSyncTopic"] = handleBrokerRemoveSyncTopicCommand
		handlers["brokerRemoveAsyncTopic"] = handleBrokerRemoveAsyncTopicCommand
		handlers["brokerAddWhitelist"] = handleBrokerAddWhitelistCommand
		handlers["brokerAddBlacklist"] = handleBrokerAddBlacklistCommand
		handlers["brokerRemoveWhitelist"] = handleBrokerRemoveWhitelistCommand
		handlers["brokerRemoveBlacklist"] = handleBrokerRemoveBlacklistCommand
		handlers["brokerAddConfigWhitelist"] = handleBrokerAddConfigWhitelistCommand
		handlers["brokerAddConfigBlacklist"] = handleBrokerAddConfigBlacklistCommand
		handlers["brokerRemoveConfigWhitelist"] = handleBrokerRemoveConfigWhitelistCommand
		handlers["brokerRemoveConfigBlacklist"] = handleBrokerRemoveConfigBlacklistCommand
	}
	if node.resolver != nil {
		handlers["resolverTopics"] = handleResolverTopicsCommand
		handlers["resolverWhitelist"] = handleResolverWhitelistCommand
		handlers["resolverBlacklist"] = handleResolverBlacklistCommand
		handlers["resolverConfigWhitelist"] = handleResolverConfigWhitelistCommand
		handlers["resolverConfigBlacklist"] = handleResolverConfigBlacklistCommand
		handlers["resolverAddWhitelist"] = handleResolverAddWhitelistCommand
		handlers["resolverAddBlacklist"] = handleResolverAddBlacklistCommand
		handlers["resolverRemoveWhitelist"] = handleResolverRemoveWhitelistCommand
		handlers["resolverRemoveBlacklist"] = handleResolverRemoveBlacklistCommand
		handlers["resolverAddConfigWhitelist"] = handleResolverAddConfigWhitelistCommand
		handlers["resolverAddConfigBlacklist"] = handleResolverAddConfigBlacklistCommand
		handlers["resolverRemoveConfigWhitelist"] = handleResolverRemoveConfigWhitelistCommand
		handlers["resolverRemoveConfigBlacklist"] = handleResolverRemoveConfigBlacklistCommand
	}
	if commandHandlerComponent := node.GetCommandHandlerComponent(); commandHandlerComponent != nil {
		commandHandlers := commandHandlerComponent.GetCommandHandlers()
		for command, commandHandler := range commandHandlers {
			handlers[command] = commandHandler
		}
	}
	return handlers
}

func handleResolverTopicsCommand(node *Node, args []string) (string, error) {
	node.resolver.mutex.Lock()
	defer node.resolver.mutex.Unlock()
	returnString := ""
	for topic, tcpEndpoint := range node.resolver.registeredTopics {
		returnString += topic + ":" + tcpEndpoint.Address + ";"
	}
	return returnString, nil
}

func handleResolverWhitelistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.resolver.resolverTcpServer.GetWhitelist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleResolverBlacklistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.resolver.resolverTcpServer.GetBlacklist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleResolverConfigWhitelistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.resolver.configTcpServer.GetWhitelist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleResolverConfigBlacklistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.resolver.configTcpServer.GetBlacklist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleResolverAddWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.resolverTcpServer.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleResolverAddBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.resolverTcpServer.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleResolverRemoveWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.resolverTcpServer.GetWhitelist().Remove(ip)
	}
	return "success", nil
}

func handleResolverRemoveBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.resolverTcpServer.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleResolverAddConfigWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.configTcpServer.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleResolverAddConfigBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.configTcpServer.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleResolverRemoveConfigWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.configTcpServer.GetWhitelist().Remove(ip)
	}
	return "success", nil
}

func handleResolverRemoveConfigBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.resolver.configTcpServer.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleBrokerNodeConnectionsCommand(node *Node, args []string) (string, error) {
	node.broker.mutex.Lock()
	defer node.broker.mutex.Unlock()
	returnString := ""
	for address := range node.broker.nodeConnections {
		returnString += address + ";"
	}
	return returnString, nil
}

func handleBrokerNodeConnectionCountCommand(node *Node, args []string) (string, error) {
	node.broker.mutex.Lock()
	defer node.broker.mutex.Unlock()
	returnString := ""
	returnString += Helpers.IntToString(len(node.broker.nodeConnections))
	return returnString, nil
}

func handleBrokerSyncTopicsCommand(node *Node, args []string) (string, error) {
	node.broker.mutex.Lock()
	defer node.broker.mutex.Unlock()
	returnString := ""
	for topic := range node.broker.syncTopics {
		returnString += topic + ";"
	}
	return returnString, nil
}

func handleBrokerAsyncTopicsCommand(node *Node, args []string) (string, error) {
	node.broker.mutex.Lock()
	defer node.broker.mutex.Unlock()
	returnString := ""
	for topic := range node.broker.asyncTopics {
		returnString += topic + ";"
	}
	return returnString, nil
}

func handleBrokerWhitelistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.broker.brokerTcpServer.GetWhitelist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleBrokerBlacklistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.broker.brokerTcpServer.GetBlacklist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleBrokerConfigWhitelistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.broker.configTcpServer.GetWhitelist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleBrokerConfigBlacklistCommand(node *Node, args []string) (string, error) {
	returnString := ""
	for _, ip := range node.broker.configTcpServer.GetBlacklist().GetElements() {
		returnString += ip + ";"
	}
	return returnString, nil
}

func handleBrokerAddSyncTopicCommand(node *Node, args []string) (string, error) {
	node.broker.addSyncTopics(args...)
	node.broker.addResolverTopicsRemotely(node.GetName(), args...)
	return "success", nil
}

func handleBrokerAddAsyncTopicCommand(node *Node, args []string) (string, error) {
	node.broker.addAsyncTopics(args...)
	node.broker.addResolverTopicsRemotely(node.GetName(), args...)
	return "success", nil
}

func handleBrokerRemoveSyncTopicCommand(node *Node, args []string) (string, error) {
	node.broker.removeSyncTopics(args...)
	node.broker.removeResolverTopicsRemotely(node.GetName(), args...)
	return "success", nil
}

func handleBrokerRemoveAsyncTopicCommand(node *Node, args []string) (string, error) {
	node.broker.removeAsyncTopics(args...)
	node.broker.removeResolverTopicsRemotely(node.GetName(), args...)
	return "success", nil
}

func handleBrokerAddWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.brokerTcpServer.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleBrokerAddBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.brokerTcpServer.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleBrokerRemoveWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.brokerTcpServer.GetWhitelist().Remove(ip)
	}
	return "success", nil
}

func handleBrokerRemoveBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.brokerTcpServer.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleBrokerAddConfigWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.configTcpServer.GetWhitelist().Add(ip)
	}
	return "success", nil
}

func handleBrokerAddConfigBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.configTcpServer.GetBlacklist().Add(ip)
	}
	return "success", nil
}

func handleBrokerRemoveConfigWhitelistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.configTcpServer.GetWhitelist().Remove(ip)
	}
	return "success", nil
}

func handleBrokerRemoveConfigBlacklistCommand(node *Node, args []string) (string, error) {
	for _, ip := range args {
		node.broker.configTcpServer.GetBlacklist().Remove(ip)
	}
	return "success", nil
}

func handleSystemgeTopicResolutionsCommand(node *Node, args []string) (string, error) {
	node.systemge.topicResolutionMutex.Lock()
	returnString := ""
	for topic, brokerConnection := range node.systemge.topicResolutions {
		returnString += topic + ":" + brokerConnection.endpoint.Address + ";"
	}
	node.systemge.topicResolutionMutex.Unlock()
	return returnString, nil
}

func handleSystemgeBrokerConnectionsCommand(node *Node, args []string) (string, error) {
	node.systemge.brokerConnectionsMutex.Lock()
	returnString := ""
	for address := range node.systemge.brokerConnections {
		returnString += address + ";"
	}
	node.systemge.brokerConnectionsMutex.Unlock()
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
