package Resolver

import (
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Server struct {
	knownBrokers     map[string]*Resolution // broker-name -> broker
	registeredTopics map[string]*Resolution // topic -> broker
	mutex            sync.Mutex

	name   string
	logger *Utilities.Logger

	isStarted bool

	resolverPort        string
	tcpListenerResolver net.Listener

	configPort        string
	tlsCertPath       string
	tlsKeyPath        string
	tlsListenerConfig net.Listener
}

func New(name, resolverPort, configPort, tlsCertPath, tlsKeyPath string, logger *Utilities.Logger) *Server {
	return &Server{
		name:             name,
		logger:           logger,
		knownBrokers:     map[string]*Resolution{},
		registeredTopics: map[string]*Resolution{},

		resolverPort: resolverPort,

		configPort:  configPort,
		tlsCertPath: tlsCertPath,
		tlsKeyPath:  tlsKeyPath,
	}
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Server already started", nil)
	}
	server.isStarted = true
	tcpListener, err := net.Listen("tcp", server.resolverPort)
	if err != nil {
		return Utilities.NewError("", err)
	}
	tlsCert, err := tls.LoadX509KeyPair(server.tlsCertPath, server.tlsKeyPath)
	if err != nil {
		return Utilities.NewError("Failed to load TLS certificate: ", err)
	}
	tlsListener, err := tls.Listen("tcp", server.configPort, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
	if err != nil {
		return Utilities.NewError("", err)
	}
	server.tcpListenerResolver = tcpListener
	server.tlsListenerConfig = tlsListener
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
	server.tcpListenerResolver.Close()
	server.tcpListenerResolver = nil
	server.tlsListenerConfig.Close()
	server.tlsListenerConfig = nil
	return nil
}

func (server *Server) IsStarted() bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.isStarted
}
