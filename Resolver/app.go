package Resolver

import (
	"Systemge/Config"
	"Systemge/Node"
	"net"
	"sync"
)

type Resolver struct {
	config *Config.Resolver
	node   *Node.Node

	tlsResolverListener net.Listener
	tlsConfigListener   net.Listener

	resolverWhitelist map[string]bool
	resolverBlacklist map[string]bool
	configWhitelist   map[string]bool
	configBlacklist   map[string]bool

	registeredTopics map[string]Config.TcpEndpoint // topic -> tcpEndpoint

	isStarted bool
	mutex     sync.Mutex
}

func New(config *Config.Resolver) *Resolver {
	resolver := &Resolver{
		config:           config,
		registeredTopics: map[string]Config.TcpEndpoint{},

		resolverWhitelist: map[string]bool{},
		resolverBlacklist: map[string]bool{},
		configWhitelist:   map[string]bool{},
		configBlacklist:   map[string]bool{},
	}
	for _, ip := range resolver.config.ResolverWhitelist {
		resolver.addToResolverWhitelist(ip)
	}
	for _, ip := range resolver.config.ConfigWhitelist {
		resolver.addToConfigWhitelist(ip)
	}
	for _, ip := range resolver.config.ResolverBlacklist {
		resolver.addToResolverBlacklist(ip)
	}
	for _, ip := range resolver.config.ConfigBlacklist {
		resolver.addToConfigBlacklist(ip)
	}
	return resolver
}
