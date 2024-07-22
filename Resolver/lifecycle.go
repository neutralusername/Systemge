package Resolver

import (
	"Systemge/Error"
	"Systemge/Node"
	"Systemge/Tcp"
)

func (resolver *Resolver) OnStart(node *Node.Node) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolver.isStarted {
		return Error.New("resolver \""+node.GetName()+"\" is already started", nil)
	}
	listener, err := Tcp.NewServer(resolver.config.Server)
	if err != nil {
		return Error.New("Failed to get listener for resolver \""+node.GetName()+"\"", err)
	}
	configListener, err := Tcp.NewServer(resolver.config.ConfigServer)
	if err != nil {
		return Error.New("Failed to get listener for resolver \""+node.GetName()+"\"", err)
	}
	resolver.resolverTcpServer = listener
	resolver.configTcpServer = configListener
	resolver.isStarted = true
	resolver.node = node

	go resolver.handleResolverConnections()
	go resolver.handleConfigConnections()
	return nil
}

func (resolver *Resolver) OnStop(node *Node.Node) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if !resolver.isStarted {
		return Error.New("resolver is not started", nil)
	}
	for topic := range resolver.registeredTopics {
		delete(resolver.registeredTopics, topic)
	}
	resolver.isStarted = false
	resolver.resolverTcpServer.GetListener().Close()
	resolver.resolverTcpServer = nil
	resolver.configTcpServer.GetListener().Close()
	resolver.configTcpServer = nil
	resolver.node = nil
	return nil
}
