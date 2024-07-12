package Resolver

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
	"net"
	"sync"
)

type Resolver struct {
	config Config.Resolver
	logger *Utilities.Logger

	registeredTopics map[string]TcpEndpoint.TcpEndpoint // topic -> tcpEndpoint

	tlsResolverListener net.Listener
	tlsConfigListener   net.Listener

	isStarted bool
	mutex     sync.Mutex
}

func New(config Config.Resolver) *Resolver {
	resolver := &Resolver{
		config:           config,
		logger:           config.Logger,
		registeredTopics: map[string]TcpEndpoint.TcpEndpoint{},
	}
	return resolver
}

func (resolver *Resolver) Start() error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolver.isStarted {
		return Error.New("resolver already started", nil)
	}
	listener, err := resolver.config.Server.GetTlsListener()
	if err != nil {
		return Error.New("Failed to get listener: ", err)
	}
	configListener, err := resolver.config.ConfigServer.GetTlsListener()
	if err != nil {
		return Error.New("Failed to get listener: ", err)
	}
	resolver.tlsResolverListener = listener
	resolver.tlsConfigListener = configListener
	resolver.isStarted = true
	go resolver.handleResolverConnections()
	go resolver.handleConfigConnections()
	resolver.config.Logger.Info("Started resolver \"" + resolver.config.Name + "\"")
	return nil
}

func (resolver *Resolver) GetName() string {
	return resolver.config.Name
}

func (resolver *Resolver) Stop() error {
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
	resolver.config.Logger.Info("Stopped resolver \"" + resolver.config.Name + "\"")
	return nil
}

func (resolver *Resolver) IsStarted() bool {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	return resolver.isStarted
}
