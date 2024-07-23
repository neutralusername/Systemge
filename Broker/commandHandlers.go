package Broker

import (
	"Systemge/Error"
	"Systemge/Node"
)

// returns a map of command handlers for the command-line interface
func (broker *Broker) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"brokerNodes": func(node *Node.Node, args []string) (string, error) {
			broker.operationMutex.Lock()
			defer broker.operationMutex.Unlock()
			resultStr := ""
			for _, nodeConnection := range broker.nodeConnections {
				resultStr += nodeConnection.name + ";"
			}
			return resultStr, nil
		},
		"syncTopics": func(node *Node.Node, args []string) (string, error) {
			broker.operationMutex.Lock()
			defer broker.operationMutex.Unlock()
			resultStr := ""
			for topic := range broker.syncTopics {
				resultStr += topic + ";"
			}
			return resultStr, nil
		},
		"asyncTopics": func(node *Node.Node, args []string) (string, error) {
			broker.operationMutex.Lock()
			defer broker.operationMutex.Unlock()
			resultStr := ""
			for topic := range broker.asyncTopics {
				resultStr += topic + ";"
			}
			return resultStr, nil
		},
		"brokerWhitelist": func(node *Node.Node, args []string) (string, error) {
			brokerWhiteList := broker.brokerTcpServer.GetWhitelist().GetElements()
			resultStr := ""
			for _, ip := range brokerWhiteList {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"brokerBlacklist": func(node *Node.Node, args []string) (string, error) {
			brokerBlackList := broker.brokerTcpServer.GetBlacklist().GetElements()
			resultStr := ""
			for _, ip := range brokerBlackList {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"configWhitelist": func(node *Node.Node, args []string) (string, error) {
			configWhiteList := broker.configTcpServer.GetWhitelist().GetElements()
			resultStr := ""
			for _, ip := range configWhiteList {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"configBlacklist": func(node *Node.Node, args []string) (string, error) {
			configBlackList := broker.configTcpServer.GetBlacklist().GetElements()
			resultStr := ""
			for _, ip := range configBlackList {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"addSyncTopic": func(node *Node.Node, args []string) (string, error) {
			broker.addSyncTopics(args...)
			broker.addResolverTopicsRemotely(args...)
			return "success", nil
		},
		"addAsyncTopic": func(node *Node.Node, args []string) (string, error) {
			broker.addAsyncTopics(args...)
			broker.addResolverTopicsRemotely(args...)
			return "success", nil
		},
		"removeSyncTopic": func(node *Node.Node, args []string) (string, error) {
			broker.removeSyncTopics(args...)
			broker.removeResolverTopicsRemotely(args...)
			return "success", nil
		},
		"removeAsyncTopic": func(node *Node.Node, args []string) (string, error) {
			broker.removeAsyncTopics(args...)
			broker.removeResolverTopicsRemotely(args...)
			return "success", nil
		},
		"addBrokerWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.brokerTcpServer.GetWhitelist().Add(ip)
			}
			return "success", nil
		},
		"addBrokerBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.brokerTcpServer.GetBlacklist().Add(ip)
			}
			return "success", nil
		},
		"removeBrokerWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.brokerTcpServer.GetWhitelist().Remove(ip)
			}
			return "success", nil
		},
		"removeBrokerBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.brokerTcpServer.GetBlacklist().Remove(ip)
			}
			return "success", nil
		},
		"addConfigWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.configTcpServer.GetWhitelist().Add(ip)
			}
			return "success", nil
		},
		"addConfigBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.configTcpServer.GetBlacklist().Add(ip)
			}
			return "success", nil
		},
		"removeConfigWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.configTcpServer.GetWhitelist().Remove(ip)
			}
			return "success", nil
		},
		"removeConfigBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				broker.configTcpServer.GetBlacklist().Remove(ip)
			}
			return "success", nil
		},
		"propagateTopics": func(node *Node.Node, args []string) (string, error) {
			if len(args) > 0 {
				for _, topic := range args {
					if err := broker.addResolverTopicsRemotely(topic); err != nil {
						return "", Error.New("Failed to add resolver topic remotely", err)
					}
				}
			} else {
				topics := []string{}
				for syncTopic := range broker.syncTopics {
					if syncTopic != "subscribe" && syncTopic != "unsubscribe" {
						topics = append(topics, syncTopic)
					}
				}
				for asyncTopic := range broker.asyncTopics {
					if asyncTopic != "heartbeat" {
						topics = append(topics, asyncTopic)
					}
				}
				err := broker.propagateTopics(topics...)
				if err != nil {
					return "", err
				}
			}
			return "success", nil
		},
	}
}
