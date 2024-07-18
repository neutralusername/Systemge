package Broker

import (
	"Systemge/Node"
)

// returns a map of command handlers for the command-line interface
func (broker *Broker) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"brokerNodes": func(node *Node.Node, args []string) error {
			broker.operationMutex.Lock()
			defer broker.operationMutex.Unlock()
			for _, nodeConnection := range broker.nodeConnections {
				println(nodeConnection.name)
			}
			return nil
		},
	}
}
