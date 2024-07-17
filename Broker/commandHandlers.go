package Broker

import "Systemge/Module"

// returns a map of command handlers for the command-line interface
func (broker *Broker) GetCommandHandlers() map[string]Module.CommandHandler {
	return map[string]Module.CommandHandler{
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
