package MessageBrokerServer

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Server struct {
	syncTopics  map[string]bool
	asyncTopics map[string]bool

	subscriptions    map[string]map[string]*Client // topic -> [client name-> client]
	connectedClients map[string]*Client            // client name -> Client
	openSyncRequests map[string]*Client            // sync request token -> client
	mutex            sync.Mutex

	name   string
	logger *Utilities.Logger

	isStarted bool

	tlsCertPath     string
	tlsKeyPath      string
	tlsListenerPort string
	tlsListener     net.Listener
}

func New(name, tlsListerPort, tlsCertPath, tlsKeyPath string, logger *Utilities.Logger) *Server {
	return &Server{
		syncTopics: map[string]bool{
			"subscribe":   true,
			"unsubscribe": true,
			"consume":     true,
		},
		asyncTopics: map[string]bool{
			"heartbeat": true,
		},

		name:   name,
		logger: logger,

		tlsCertPath:     tlsCertPath,
		tlsKeyPath:      tlsKeyPath,
		tlsListenerPort: tlsListerPort,
	}
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Error.New("Server already started", nil)
	}
	var tlsListener net.Listener
	if server.tlsCertPath == "" || server.tlsKeyPath == "" || server.tlsListenerPort == "" {
		return Error.New("TLS certificate path, TLS key path, and TLS listener port must be provided", nil)
	}
	tlsCert, err := tls.LoadX509KeyPair(server.tlsCertPath, server.tlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	tlsListener, err = tls.Listen("tcp", server.tlsListenerPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{tlsCert},
	})
	if err != nil {
		return Error.New("Failed to start server: ", err)
	}
	server.connectedClients = map[string]*Client{}
	server.openSyncRequests = map[string]*Client{}
	server.subscriptions = map[string]map[string]*Client{}
	server.tlsListener = tlsListener
	go server.handleTlsConnections()
	server.isStarted = true
	return nil
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) Stop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.isStarted {
		return Error.New("Server is not started", nil)
	}
	server.isStarted = false
	for _, clients := range server.subscriptions {
		for _, client := range clients {
			delete(server.connectedClients, client.name)
		}
	}
	server.tlsListener.Close()
	server.tlsListener = nil
	return nil
}

func (server *Server) IsStarted() bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.isStarted
}
