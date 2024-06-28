package Resolver

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Resolver struct {
	config *Config.Resolver
	logger *Utilities.Logger

	knownBrokers     map[string]*knownBroker // broker-name -> broker
	registeredTopics map[string]*knownBroker // topic -> broker

	tlsResolverListener net.Listener
	tlsConfigListener   net.Listener

	isStarted bool
	mutex     sync.Mutex
}

func New(config *Config.Resolver) *Resolver {
	return &Resolver{
		config:           config,
		logger:           Utilities.NewLogger(config.LoggerPath),
		knownBrokers:     map[string]*knownBroker{},
		registeredTopics: map[string]*knownBroker{},
	}
}

func (resolver *Resolver) Start() error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolver.isStarted {
		return Error.New("resolver already started", nil)
	}
	resolverTlsCert, err := tls.LoadX509KeyPair(resolver.config.ResolverTlsCertPath, resolver.config.ResolverTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	resolverListener, err := tls.Listen("tcp", resolver.config.ResolverPort, &tls.Config{
		Certificates: []tls.Certificate{resolverTlsCert},
	})
	if err != nil {
		return Error.New("Failed to listen on port: ", err)
	}
	configTlsCert, err := tls.LoadX509KeyPair(resolver.config.ConfigTlsCertPath, resolver.config.ConfigTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	configListener, err := tls.Listen("tcp", resolver.config.ConfigPort, &tls.Config{
		Certificates: []tls.Certificate{configTlsCert},
	})
	if err != nil {
		return Error.New("Failed to listen on port: ", err)
	}
	resolver.tlsResolverListener = resolverListener
	resolver.tlsConfigListener = configListener
	resolver.isStarted = true
	go resolver.handleResolverConnections()
	go resolver.handleConfigConnections()
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
	return nil
}

func (resolver *Resolver) IsStarted() bool {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	return resolver.isStarted
}
