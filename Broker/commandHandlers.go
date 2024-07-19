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
		"syncTopics": func(node *Node.Node, args []string) error {
			broker.operationMutex.Lock()
			defer broker.operationMutex.Unlock()
			for topic := range broker.syncTopics {
				println(topic)
			}
			return nil
		},
		"asyncTopics": func(node *Node.Node, args []string) error {
			broker.operationMutex.Lock()
			defer broker.operationMutex.Unlock()
			for topic := range broker.asyncTopics {
				println(topic)
			}
			return nil
		},
		"whitelist": func(node *Node.Node, args []string) error {
			broker.stateMutex.Lock()
			defer broker.stateMutex.Unlock()
			for ip := range broker.whitelist {
				println(ip)
			}
			return nil
		},
		"blacklist": func(node *Node.Node, args []string) error {
			broker.stateMutex.Lock()
			defer broker.stateMutex.Unlock()
			for ip := range broker.blacklist {
				println(ip)
			}
			return nil
		},
		"addSyncTopic": func(node *Node.Node, args []string) error {
			broker.addSyncTopics(args...)
			broker.addResolverTopicsRemotely(args...)
			return nil
		},
		"addAsyncTopic": func(node *Node.Node, args []string) error {
			broker.addAsyncTopics(args...)
			broker.addResolverTopicsRemotely(args...)
			return nil
		},
		"removeSyncTopic": func(node *Node.Node, args []string) error {
			broker.removeSyncTopics(args...)
			broker.removeResolverTopicsRemotely(args...)
			return nil
		},
		"removeAsyncTopic": func(node *Node.Node, args []string) error {
			broker.removeAsyncTopics(args...)
			broker.removeResolverTopicsRemotely(args...)
			return nil
		},
		"addWhitelist": func(node *Node.Node, args []string) error {
			broker.addToWhitelist(args...)
			return nil
		},
		"addBlacklist": func(node *Node.Node, args []string) error {
			broker.addToBlacklist(args...)
			return nil
		},
		"removeWhitelist": func(node *Node.Node, args []string) error {
			broker.removeFromWhitelist(args...)
			return nil
		},
		"removeBlacklist": func(node *Node.Node, args []string) error {
			broker.removeFromBlacklist(args...)
			return nil
		},
	}
}
