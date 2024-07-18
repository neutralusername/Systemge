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

	registeredTopics map[string]Config.TcpEndpoint // topic -> tcpEndpoint

	isStarted bool
	mutex     sync.Mutex
}

func New(config *Config.Resolver) *Resolver {
	resolver := &Resolver{
		config:           config,
		registeredTopics: map[string]Config.TcpEndpoint{},
	}
	return resolver
}
