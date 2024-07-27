package Resolver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tcp"
)

type Resolver struct {
	config *Config.Resolver
	node   *Node.Node

	resolverTcpServer *Tcp.Server
	configTcpServer   *Tcp.Server

	registeredTopics map[string]Config.TcpEndpoint // topic -> tcpEndpoint

	isStarted bool
	mutex     sync.Mutex

	configRequestCounter     atomic.Uint32
	resolutionRequestCounter atomic.Uint32
	bytesReceivedCounter     atomic.Uint64
	bytesSentCounter         atomic.Uint64
}

func (resolver *Resolver) RetrieveConfigRequestCounter() uint32 {
	return resolver.configRequestCounter.Swap(0)
}

func (resolver *Resolver) RetrieveResolutionRequestCounter() uint32 {
	return resolver.resolutionRequestCounter.Swap(0)
}

func (resolver *Resolver) RetrieveBytesReceivedCounter() uint64 {
	return resolver.bytesReceivedCounter.Swap(0)
}

func (resolver *Resolver) RetrieveBytesSentCounter() uint64 {
	return resolver.bytesSentCounter.Swap(0)
}

func ImplementsResolver(application Node.Application) bool {
	_, ok := application.(*Resolver)
	return ok
}

func New(config *Config.Resolver) *Resolver {
	resolver := &Resolver{
		config:           config,
		registeredTopics: map[string]Config.TcpEndpoint{},
	}
	return resolver
}
