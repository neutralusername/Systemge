package Broker

import "Systemge/Module"

// returns a map of custom command handlers for the command-line interface
func (broker *Broker) GetCustomCommandHandlers() map[string]Module.CustomCommandHandler {
	return map[string]Module.CustomCommandHandler{
		"brokerNodes": func(args []string) error {
			broker.operationMutex.Lock()
			for _, nodeConnection := range broker.nodeConnections {
				println(nodeConnection.name)
			}
			broker.operationMutex.Unlock()
			return nil
		},
	}
}
