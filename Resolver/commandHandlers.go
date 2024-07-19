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
			for _, ip := range resolver.resolverWhitelist.GetElements() {
				println(ip)
			}
			return nil
		},
		"resolverBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range resolver.resolverBlacklist.GetElements() {
				println(ip)
			}
			return nil
		},
		"configWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range resolver.configWhitelist.GetElements() {
				println(ip)
			}
			return nil
		},
		"configBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range resolver.configBlacklist.GetElements() {
				println(ip)
			}
			return nil
		},
		"addResolverWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.resolverWhitelist.Add(ip)
			}
			return nil
		},
		"addResolverBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.resolverBlacklist.Add(ip)
			}
			return nil
		},
		"removeResolverWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.resolverWhitelist.Remove(ip)
			}
			return nil
		},
		"removeResolverBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.resolverBlacklist.Remove(ip)
			}
			return nil
		},
		"addConfigWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.configWhitelist.Add(ip)
			}
			return nil
		},
		"addConfigBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.configBlacklist.Add(ip)
			}
			return nil
		},
		"removeConfigWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.configWhitelist.Remove(ip)
			}
			return nil
		},
		"removeConfigBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				resolver.configBlacklist.Remove(ip)
			}
			return nil
		},
	}
}
