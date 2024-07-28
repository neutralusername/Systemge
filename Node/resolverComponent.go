package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
)

type resolverComponent struct {
	application ResolverComponent

	resolverTcpServer *Tcp.Server
	configTcpServer   *Tcp.Server

	registeredTopics map[string]Config.TcpEndpoint // topic -> tcpEndpoint

	mutex sync.Mutex

	configRequestCounter     atomic.Uint32
	resolutionRequestCounter atomic.Uint32
	bytesReceivedCounter     atomic.Uint64
	bytesSentCounter         atomic.Uint64
}

func (node *Node) RetrieveResolverConfigRequestCounter() uint32 {
	if resolver := node.resolver; resolver != nil {
		return resolver.configRequestCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveResolverResolutionRequestCounter() uint32 {
	if resolver := node.resolver; resolver != nil {
		return resolver.resolutionRequestCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveResolverBytesReceivedCounter() uint64 {
	if resolver := node.resolver; resolver != nil {
		return resolver.bytesReceivedCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveResolverBytesSentCounter() uint64 {
	if resolver := node.resolver; resolver != nil {
		return resolver.bytesSentCounter.Swap(0)
	}
	return 0
}

func (node *Node) startResolverComponent() error {
	node.resolver = &resolverComponent{
		application:      node.application.(ResolverComponent),
		registeredTopics: map[string]Config.TcpEndpoint{},
	}
	listener, err := Tcp.NewServer(node.resolver.application.GetResolverComponentConfig().Server)
	if err != nil {
		return Error.New("Failed to get listener for resolver \""+node.GetName()+"\"", err)
	}
	configListener, err := Tcp.NewServer(node.resolver.application.GetResolverComponentConfig().ConfigServer)
	if err != nil {
		return Error.New("Failed to get listener for resolver \""+node.GetName()+"\"", err)
	}
	node.resolver.resolverTcpServer = listener
	node.resolver.configTcpServer = configListener

	go node.handleResolverConnections()
	go node.handleResolverConfigConnections()
	return nil
}

func (node *Node) stopResolverComponent() error {
	node.resolver.mutex.Lock()
	defer node.resolver.mutex.Unlock()
	for topic := range node.resolver.registeredTopics {
		delete(node.resolver.registeredTopics, topic)
	}
	node.resolver.resolverTcpServer.GetListener().Close()
	node.resolver.resolverTcpServer = nil
	node.resolver.configTcpServer.GetListener().Close()
	node.resolver.configTcpServer = nil
	return nil
}
