package Resolver

import (
	"Systemge/Node"
)

// returns a map of custom command handlers for the command-line interface
func (resolver *Resolver) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"resolverTopics": resolver.handleTopicsCommand,
	}
}

func (resolver *Resolver) handleTopicsCommand(node *Node.Node, args []string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for topic, tcpEndpoint := range resolver.registeredTopics {
		println(topic, tcpEndpoint.Address)
	}
	return nil
}
