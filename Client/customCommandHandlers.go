package Client

import (
	"Systemge/Application"
	"Systemge/Utilities"
)

// returns a map of custom command handlers for the command-line interface
func (client *Client) GetCustomCommandHandlers() map[string]Application.CustomCommandHandler {
	return map[string]Application.CustomCommandHandler{
		"brokers":          client.handleBrokersCommand,
		"removeBroker":     client.handleRemoveBrokerCommand,
		"resolutions":      client.handleResolutionsCommand,
		"removeResolution": client.handleRemoveTopicCommand,
	}
}

func (client *Client) handleBrokersCommand(args []string) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
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
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
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
