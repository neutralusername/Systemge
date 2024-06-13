package MessageBrokerServer

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"net"
	"sync"
)

type Server struct {
	subscriptions    map[string]map[string]*Client // topic -> [client name-> client]
	clients          map[string]*Client            // client name -> Client
	openSyncRequests map[string]*Client            // sync request token -> client
	mutex            sync.Mutex

	name         string
	listenerPort string
	logger       *Utilities.Logger

	isStarted bool

	tcpListener net.Listener
}

func New(name string, listenerPort string, logger *Utilities.Logger) *Server {
	return &Server{
		subscriptions:    map[string]map[string]*Client{},
		clients:          nil,
		openSyncRequests: nil,

		mutex: sync.Mutex{},

		name:         name,
		listenerPort: listenerPort,
		logger:       logger,

		isStarted: false,

		tcpListener: nil,
	}
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Error.New("Server already started", nil)
	}
	tcpListener, err := net.Listen("tcp", server.listenerPort)
	if err != nil {
		return Error.New("Failed to start server: ", err)
	}
	server.clients = map[string]*Client{}
	server.openSyncRequests = map[string]*Client{}
	server.tcpListener = tcpListener
	server.isStarted = true
	go server.handleConnections()
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
			delete(server.clients, client.name)
		}
	}
	server.tcpListener.Close()
	server.tcpListener = nil
	return nil
}

func (server *Server) IsStarted() bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.isStarted
}
