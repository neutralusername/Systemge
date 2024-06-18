package Broker

import (
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Server struct {
	syncTopics  map[string]bool
	asyncTopics map[string]bool

	clientSubscriptions map[string]map[string]*clientConnection // topic -> [clientName-> client]
	clientConnections   map[string]*clientConnection            // clientName -> Client
	openSyncRequests    map[string]*syncRequest

	operationMutex sync.Mutex
	stateMutex     sync.Mutex

	name   string
	logger *Utilities.Logger

	isStarted bool

	brokerTlsCertPath string
	brokerTlsKeyPath  string
	brokerPort        string
	tlsBrokerListener net.Listener

	configTlsCertPath string
	configTlsKeyPath  string
	configPort        string
	tlsConfigListener net.Listener
}

func New(name, brokerPort, brokerTlsCertPath, brokerTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath string, logger *Utilities.Logger) *Server {
	return &Server{
		syncTopics: map[string]bool{
			"subscribe":   true,
			"unsubscribe": true,
			"consume":     true,
		},
		asyncTopics: map[string]bool{
			"heartbeat": true,
		},

		clientSubscriptions: map[string]map[string]*clientConnection{},
		clientConnections:   map[string]*clientConnection{},
		openSyncRequests:    map[string]*syncRequest{},

		name:   name,
		logger: logger,

		brokerTlsCertPath: brokerTlsCertPath,
		brokerTlsKeyPath:  brokerTlsKeyPath,
		brokerPort:        brokerPort,

		configTlsCertPath: configTlsCertPath,
		configTlsKeyPath:  configTlsKeyPath,
		configPort:        configPort,
	}
}

func (server *Server) Start() error {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Server already started", nil)
	}
	if server.brokerTlsCertPath == "" || server.brokerTlsKeyPath == "" || server.brokerPort == "" {
		return Utilities.NewError("TLS certificate path, TLS key path, and TLS listener port must be provided", nil)
	}
	brokerCert, err := tls.LoadX509KeyPair(server.brokerTlsCertPath, server.brokerTlsKeyPath)
	if err != nil {
		return Utilities.NewError("Failed to load TLS certificate: ", err)
	}
	brokerListener, err := tls.Listen("tcp", server.brokerPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{brokerCert},
	})
	if err != nil {
		return Utilities.NewError("Failed to start server: ", err)
	}
	configCert, err := tls.LoadX509KeyPair(server.configTlsCertPath, server.configTlsKeyPath)
	if err != nil {
		return Utilities.NewError("Failed to load TLS certificate: ", err)
	}
	configListener, err := tls.Listen("tcp", server.configPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{configCert},
	})
	if err != nil {
		return Utilities.NewError("Failed to start server: ", err)
	}
	server.tlsBrokerListener = brokerListener
	server.tlsConfigListener = configListener
	server.isStarted = true
	go server.handleBrokerConnections()
	go server.handleConfigConnections()
	return nil
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) Stop() error {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if !server.isStarted {
		return Utilities.NewError("Server is not started", nil)
	}
	clientsToDisconnect := make([]*clientConnection, 0)
	server.operationMutex.Lock()
	for _, clientConnection := range server.clientConnections {
		clientsToDisconnect = append(clientsToDisconnect, clientConnection)
	}
	server.operationMutex.Unlock()
	for _, clientConnection := range clientsToDisconnect {
		clientConnection.disconnect()

	}

	server.tlsBrokerListener.Close()
	server.tlsConfigListener.Close()
	server.isStarted = false
	return nil
}
