package Resolver

import (
	"github.com/neutralusername/Systemge/Node"
)

// returns a map of custom command handlers for the command-line interface
func (resolver *Resolver) GetCommandHandlers() map[string]Node.CommandHandler {
	return map[string]Node.CommandHandler{
		"resolverTopics": func(node *Node.Node, args []string) (string, error) {
			resolver.mutex.Lock()
			defer resolver.mutex.Unlock()
			resultStr := ""
			for topic, tcpEndpoint := range resolver.registeredTopics {
				resultStr += topic + ":" + tcpEndpoint.Address + ";"
			}
			return resultStr, nil
		},
		"resolverWhitelist": func(node *Node.Node, args []string) (string, error) {
			resultStr := ""
			for _, ip := range resolver.resolverTcpServer.GetWhitelist().GetElements() {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"resolverBlacklist": func(node *Node.Node, args []string) (string, error) {
			resultStr := ""
			for _, ip := range resolver.resolverTcpServer.GetBlacklist().GetElements() {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"configWhitelist": func(node *Node.Node, args []string) (string, error) {
			resultStr := ""
			for _, ip := range resolver.configTcpServer.GetWhitelist().GetElements() {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"configBlacklist": func(node *Node.Node, args []string) (string, error) {
			resultStr := ""
			for _, ip := range resolver.configTcpServer.GetBlacklist().GetElements() {
				resultStr += ip + ";"
			}
			return resultStr, nil
		},
		"addResolverWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.resolverTcpServer.GetWhitelist().Add(ip)
			}
			return "success", nil
		},
		"addResolverBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.resolverTcpServer.GetBlacklist().Add(ip)
			}
			return "success", nil
		},
		"removeResolverWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.resolverTcpServer.GetWhitelist().Remove(ip)
			}
			return "success", nil
		},
		"removeResolverBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.resolverTcpServer.GetBlacklist().Remove(ip)
			}
			return "success", nil
		},
		"addConfigWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.configTcpServer.GetWhitelist().Add(ip)
			}
			return "success", nil
		},
		"addConfigBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.configTcpServer.GetBlacklist().Add(ip)
			}
			return "success", nil
		},
		"removeConfigWhitelist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.configTcpServer.GetWhitelist().Remove(ip)
			}
			return "success", nil
		},
		"removeConfigBlacklist": func(node *Node.Node, args []string) (string, error) {
			for _, ip := range args {
				resolver.configTcpServer.GetBlacklist().Remove(ip)
			}
			return "success", nil
		},
	}
}
