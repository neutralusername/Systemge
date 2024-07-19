package Resolver

import (
	"Systemge/Node"
)

// returns a map of custom command handlers for the command-line interface
func (resolver *Resolver) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"resolverTopics": func(node *Node.Node, args []string) error {
			resolver.mutex.Lock()
			defer resolver.mutex.Unlock()
			for topic, tcpEndpoint := range resolver.registeredTopics {
				println(topic, tcpEndpoint.Address)
			}
			return nil
		},
		"resolverWhitelist": func(node *Node.Node, args []string) error {
			resolver.mutex.Lock()
			defer resolver.mutex.Unlock()
			for ip := range resolver.resolverWhitelist {
				println(ip)
			}
			return nil
		},
		"resolverBlacklist": func(node *Node.Node, args []string) error {
			resolver.mutex.Lock()
			defer resolver.mutex.Unlock()
			for ip := range resolver.resolverBlacklist {
				println(ip)
			}
			return nil
		},
		"configWhitelist": func(node *Node.Node, args []string) error {
			resolver.mutex.Lock()
			defer resolver.mutex.Unlock()
			for ip := range resolver.configWhitelist {
				println(ip)
			}
			return nil
		},
		"configBlacklist": func(node *Node.Node, args []string) error {
			resolver.mutex.Lock()
			defer resolver.mutex.Unlock()
			for ip := range resolver.configBlacklist {
				println(ip)
			}
			return nil
		},
		"addResolverWhitelist": func(node *Node.Node, args []string) error {
			resolver.addToresolverWhitelist(args...)
			return nil
		},
		"addResolverBlacklist": func(node *Node.Node, args []string) error {
			resolver.addToresolverBlacklist(args...)
			return nil
		},
		"removeResolverWhitelist": func(node *Node.Node, args []string) error {
			resolver.removeFromresolverWhitelist(args...)
			return nil
		},
		"removeResolverBlacklist": func(node *Node.Node, args []string) error {
			resolver.removeFromresolverBlacklist(args...)
			return nil
		},
		"addConfigWhitelist": func(node *Node.Node, args []string) error {
			resolver.addToConfigWhitelist(args...)
			return nil
		},
		"addConfigBlacklist": func(node *Node.Node, args []string) error {
			resolver.addToConfigBlacklist(args...)
			return nil
		},
		"removeConfigWhitelist": func(node *Node.Node, args []string) error {
			resolver.removeFromConfigWhitelist(args...)
			return nil
		},
		"removeConfigBlacklist": func(node *Node.Node, args []string) error {
			resolver.removeFromConfigBlacklist(args...)
			return nil
		},
	}
}
