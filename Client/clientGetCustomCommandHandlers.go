package Client

import (
	"Systemge/Utilities"
)

// returns a map of custom command handlers for the command-line interface
func (client *Client) GetCustomCommandHandlers() map[string]func([]string) error {
	handlers := map[string]func([]string) error{
		"brokers":          client.handleBrokersCommand,
		"removeBroker":     client.handleRemoveBrokerCommand,
		"resolutions":      client.handleResolutionsCommand,
		"removeResolution": client.handleRemoveTopicCommand,
		"websocketClients": client.handleWebsocketClientsCommand,
		"groups":           client.handleGroupsCommand,
		"groupClients":     client.handleGroupClientsCommand,
	}
	if client.application != nil {
		customHandlers := client.application.GetCustomCommandHandlers()
		for command, handler := range customHandlers {
			handlers[command] = func(args []string) error {
				return handler(client, args)
			}
		}
	}
	return handlers
}

func (client *Client) handleBrokersCommand(args []string) error {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	for _, brokerConnection := range client.activeBrokerConnections {
		println(brokerConnection.resolution.GetName() + " : " + brokerConnection.resolution.GetAddress())
	}
	return nil
}

func (client *Client) handleRemoveBrokerCommand(args []string) error {
	if len(args) != 1 {
		return Utilities.NewError("Invalid number of arguments", nil)
	}
	brokerAddress := args[0]
	err := client.RemoveBrokerConnection(brokerAddress)
	if err != nil {
		return Utilities.NewError("Error removing broker connection", err)
	}
	return nil
}

func (client *Client) handleResolutionsCommand(args []string) error {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	for topic, brokerConnection := range client.topicResolutions {
		println(topic + " : " + brokerConnection.resolution.GetName() + " : " + brokerConnection.resolution.GetAddress())
	}
	return nil
}

func (client *Client) handleRemoveTopicCommand(args []string) error {
	if len(args) != 1 {
		return Utilities.NewError("Invalid number of arguments", nil)
	}
	topic := args[0]
	err := client.RemoveTopicResolution(topic)
	if err != nil {
		return Utilities.NewError("Error removing topic resolution", err)
	}
	return nil
}

func (client *Client) handleWebsocketClientsCommand(args []string) error {
	client.websocketMutex.Lock()
	for _, client := range client.websocketClients {
		println(client.GetId())
	}
	client.websocketMutex.Unlock()
	return nil
}

func (client *Client) handleGroupsCommand(args []string) error {
	client.websocketMutex.Lock()
	for groupId := range client.groups {
		println(groupId)
	}
	client.websocketMutex.Unlock()
	return nil
}

func (client *Client) handleGroupClientsCommand(args []string) error {
	if len(args) < 1 {
		println("Usage: groupClients <groupId>")
	}
	groupId := args[0]
	client.websocketMutex.Lock()
	group, ok := client.groups[groupId]
	client.websocketMutex.Unlock()
	if !ok {
		println("Group not found")
	} else {
		for _, client := range group {
			println(client.GetId())
		}
	}
	return nil
}
