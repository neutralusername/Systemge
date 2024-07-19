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
			brokerWhiteList := broker.brokerWhitelist.GetElements()
			for _, ip := range brokerWhiteList {
				println(ip)
			}
			return nil
		},
		"brokerBlacklist": func(node *Node.Node, args []string) error {
			brokerBlackList := broker.brokerBlacklist.GetElements()
			for _, ip := range brokerBlackList {
				println(ip)
			}
			return nil
		},
		"configWhitelist": func(node *Node.Node, args []string) error {
			configWhiteList := broker.configWhitelist.GetElements()
			for _, ip := range configWhiteList {
				println(ip)
			}
			return nil
		},
		"configBlacklist": func(node *Node.Node, args []string) error {
			configBlackList := broker.configBlacklist.GetElements()
			for _, ip := range configBlackList {
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
			for _, ip := range args {
				broker.brokerWhitelist.Add(ip)
			}
			return nil
		},
		"addBrokerBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.brokerBlacklist.Add(ip)
			}
			return nil
		},
		"removeBrokerWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.brokerWhitelist.Remove(ip)
			}
			return nil
		},
		"removeBrokerBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.brokerBlacklist.Remove(ip)
			}
			return nil
		},
		"addConfigWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.configWhitelist.Add(ip)
			}
			return nil
		},
		"addConfigBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.configBlacklist.Add(ip)
			}
			return nil
		},
		"removeConfigWhitelist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.configWhitelist.Remove(ip)
			}
			return nil
		},
		"removeConfigBlacklist": func(node *Node.Node, args []string) error {
			for _, ip := range args {
				broker.configBlacklist.Remove(ip)
			}
			return nil
		},
	}
}
