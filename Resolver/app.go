package Resolver

import (
	"Systemge/Config"
	"Systemge/Node"
	"Systemge/Tcp"
	"sync"
)

type Resolver struct {
	config *Config.Resolver
	node   *Node.Node

	resolverTcpServer *Tcp.Server
	configTcpServer   *Tcp.Server

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
