package Resolver

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Resolver struct {
	name   string
	logger *Utilities.Logger

	knownBrokers     map[string]*knownBroker // broker-name -> broker
	registeredTopics map[string]*knownBroker // topic -> broker

	resolverPort        string
	resolverTlsCertPath string
	resolverTlsKeyPath  string
	tlsResolverListener net.Listener

	configPort        string
	configTlsCertPath string
	configTlsKeyPath  string
	tlsConfigListener net.Listener

	isStarted bool
	mutex     sync.Mutex
}

func New(name, resolverPort, resolverTlsCertPath, resolverTlsKeyPath, configPort, tlsCertPath, tlsKeyPath string, logger *Utilities.Logger) *Resolver {
	return &Resolver{
		name:             name,
		logger:           logger,
		knownBrokers:     map[string]*knownBroker{},
		registeredTopics: map[string]*knownBroker{},

		resolverPort:        resolverPort,
		resolverTlsCertPath: resolverTlsCertPath,
		resolverTlsKeyPath:  resolverTlsKeyPath,

		configPort:        configPort,
		configTlsCertPath: tlsCertPath,
		configTlsKeyPath:  tlsKeyPath,
	}
}

func (resolver *Resolver) Start() error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolver.isStarted {
		return Error.New("resolver already started", nil)
	}
	resolverTlsCert, err := tls.LoadX509KeyPair(resolver.resolverTlsCertPath, resolver.resolverTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	resolverListener, err := tls.Listen("tcp", resolver.resolverPort, &tls.Config{
		Certificates: []tls.Certificate{resolverTlsCert},
	})
	if err != nil {
		return Error.New("Failed to listen on port: ", err)
	}
	configTlsCert, err := tls.LoadX509KeyPair(resolver.configTlsCertPath, resolver.configTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	configListener, err := tls.Listen("tcp", resolver.configPort, &tls.Config{
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
	return resolver.name
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
