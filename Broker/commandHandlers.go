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
		"brokerWhitelist": func(node *Node.Node, args []string) error {
			broker.stateMutex.Lock()
			defer broker.stateMutex.Unlock()
			for ip := range broker.brokerWhitelist {
				println(ip)
			}
			return nil
		},
		"brokerBlacklist": func(node *Node.Node, args []string) error {
			broker.stateMutex.Lock()
			defer broker.stateMutex.Unlock()
			for ip := range broker.brokerBlacklist {
				println(ip)
			}
			return nil
		},
		"configWhitelist": func(node *Node.Node, args []string) error {
			broker.stateMutex.Lock()
			defer broker.stateMutex.Unlock()
			for ip := range broker.configWhitelist {
				println(ip)
			}
			return nil
		},
		"configBlacklist": func(node *Node.Node, args []string) error {
			broker.stateMutex.Lock()
			defer broker.stateMutex.Unlock()
			for ip := range broker.configBlacklist {
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
		"addBrokerWhitelist": func(node *Node.Node, args []string) error {
			broker.addToBrokerWhitelist(args...)
			return nil
		},
		"addBrokerBlacklist": func(node *Node.Node, args []string) error {
			broker.addToBrokerBlacklist(args...)
			return nil
		},
		"removeBrokerWhitelist": func(node *Node.Node, args []string) error {
			broker.removeFromBrokerWhitelist(args...)
			return nil
		},
		"removeBrokerBlacklist": func(node *Node.Node, args []string) error {
			broker.removeFromBrokerBlacklist(args...)
			return nil
		},
		"addConfigWhitelist": func(node *Node.Node, args []string) error {
			broker.addToConfigWhitelist(args...)
			return nil
		},
		"addConfigBlacklist": func(node *Node.Node, args []string) error {
			broker.addToConfigBlacklist(args...)
			return nil
		},
		"removeConfigWhitelist": func(node *Node.Node, args []string) error {
			broker.removeFromConfigWhitelist(args...)
			return nil
		},
		"removeConfigBlacklist": func(node *Node.Node, args []string) error {
			broker.removeFromConfigBlacklist(args...)
			return nil
		},
	}
}
