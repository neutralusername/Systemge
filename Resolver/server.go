package Resolver

import (
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Server struct {
	knownBrokers     map[string]*knownBroker // broker-name -> broker
	registeredTopics map[string]*knownBroker // topic -> broker
	mutex            sync.Mutex

	name   string
	logger *Utilities.Logger

	isStarted bool

	resolverPort        string
	resolverTlsCertPath string
	resolverTlsKeyPath  string
	tlsResolverListener net.Listener

	configPort        string
	configTlsCertPath string
	configTlsKeyPath  string
	tlsConfigListener net.Listener
}

func New(name, resolverPort, resolverTlsCertPath, resolverTlsKeyPath, configPort, tlsCertPath, tlsKeyPath string, logger *Utilities.Logger) *Server {
	return &Server{
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

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Server already started", nil)
	}
	server.isStarted = true
	resolverTlsCert, err := tls.LoadX509KeyPair(server.resolverTlsCertPath, server.resolverTlsKeyPath)
	if err != nil {
		return Utilities.NewError("Failed to load TLS certificate: ", err)
	}
	resolverListener, err := tls.Listen("tcp", server.resolverPort, &tls.Config{
		Certificates: []tls.Certificate{resolverTlsCert},
	})
	if err != nil {
		return Utilities.NewError("", err)
	}
	configTlsCert, err := tls.LoadX509KeyPair(server.configTlsCertPath, server.configTlsKeyPath)
	if err != nil {
		return Utilities.NewError("Failed to load TLS certificate: ", err)
	}
	configListener, err := tls.Listen("tcp", server.configPort, &tls.Config{
		Certificates: []tls.Certificate{configTlsCert},
	})
	if err != nil {
		return Utilities.NewError("", err)
	}
	server.tlsResolverListener = resolverListener
	server.tlsConfigListener = configListener
	go server.handleResolverConnections()
	go server.handleConfigConnections()
	return nil
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) Stop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.isStarted {
		return Utilities.NewError("Server is not started", nil)
	}
	server.isStarted = false
	server.tlsResolverListener.Close()
	server.tlsResolverListener = nil
	server.tlsConfigListener.Close()
	server.tlsConfigListener = nil
	return nil
}

func (server *Server) IsStarted() bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.isStarted
}
