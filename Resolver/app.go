package Resolver

import (
	"Systemge/Config"
	"Systemge/Node"
	"Systemge/Tools"
	"net"
	"sync"
)

type Resolver struct {
	config *Config.Resolver
	node   *Node.Node

	tlsResolverListener net.Listener
	tlsConfigListener   net.Listener

	resolverWhitelist *Tools.AccessControlList_
	resolverBlacklist *Tools.AccessControlList_
	configWhitelist   *Tools.AccessControlList_
	configBlacklist   *Tools.AccessControlList_

	registeredTopics map[string]Config.TcpEndpoint // topic -> tcpEndpoint

	isStarted bool
	mutex     sync.Mutex
}

func New(config *Config.Resolver) *Resolver {
	resolver := &Resolver{
		config:           config,
		registeredTopics: map[string]Config.TcpEndpoint{},

		resolverWhitelist: Tools.NewAccessControlList(),
		resolverBlacklist: Tools.NewAccessControlList(),
		configWhitelist:   Tools.NewAccessControlList(),
		configBlacklist:   Tools.NewAccessControlList(),
	}
	for _, ip := range resolver.config.ResolverWhitelist {
		resolver.resolverWhitelist.Add(ip)
	}
	for _, ip := range resolver.config.ConfigWhitelist {
		resolver.configWhitelist.Add(ip)
	}
	for _, ip := range resolver.config.ResolverBlacklist {
		resolver.resolverBlacklist.Add(ip)
	}
	for _, ip := range resolver.config.ConfigBlacklist {
		resolver.configBlacklist.Add(ip)
	}
	return resolver
}
