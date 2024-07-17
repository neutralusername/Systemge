package Resolver

import (
	"Systemge/Config"
	"Systemge/Node"
	"Systemge/TcpEndpoint"
	"net"
	"sync"
)

type Resolver struct {
	config Config.Resolver
	node   *Node.Node

	tlsResolverListener net.Listener
	tlsConfigListener   net.Listener

	registeredTopics map[string]TcpEndpoint.TcpEndpoint // topic -> tcpEndpoint

	isStarted bool
	mutex     sync.Mutex
}

func New(config Config.Resolver) *Resolver {
	resolver := &Resolver{
		config:           config,
		registeredTopics: map[string]TcpEndpoint.TcpEndpoint{},
	}
	return resolver
}
