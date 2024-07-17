package Resolver

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (resolver *Resolver) OnStart(node *Node.Node) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolver.isStarted {
		return Error.New("resolver \""+node.GetName()+"\" is already started", nil)
	}
	listener, err := resolver.config.Server.GetListener()
	if err != nil {
		return Error.New("Failed to get listener for resolver \""+node.GetName()+"\"", err)
	}
	configListener, err := resolver.config.ConfigServer.GetListener()
	if err != nil {
		return Error.New("Failed to get listener for resolver \""+node.GetName()+"\"", err)
	}
	resolver.tlsResolverListener = listener
	resolver.tlsConfigListener = configListener
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
	resolver.isStarted = false
	resolver.tlsResolverListener.Close()
	resolver.tlsResolverListener = nil
	resolver.tlsConfigListener.Close()
	resolver.tlsConfigListener = nil
	resolver.node = nil
	return nil
}
