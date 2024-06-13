package TopicResolutionServer

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"net"
	"sync"
)

type Server struct {
	registeredTopics map[string]string // topic -> message-broker-server-address
	mutex            sync.Mutex

	name         string
	listenerPort string
	logger       *Utilities.Logger

	isStarted   bool
	tcpListener net.Listener
}

func New(name string, listenerPort string, logger *Utilities.Logger) *Server {
	return &Server{
		registeredTopics: make(map[string]string),
		name:             name,
		listenerPort:     listenerPort,
		logger:           logger,
	}
}

func (server *Server) RegisterTopics(messageBrokerAddress string, topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.registeredTopics[topic] = messageBrokerAddress
	}
}

func (server *Server) UnregisterTopic(topic string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.registeredTopics, topic)
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Error.New("Server already started", nil)
	}
	server.isStarted = true
	tcpListener, err := net.Listen("tcp", server.listenerPort)
	if err != nil {
		return Error.New("", err)
	}
	server.tcpListener = tcpListener
	go func() {
		for server.IsStarted() {
			netConn, err := server.tcpListener.Accept()
			if err != nil {
				if server.IsStarted() {
					server.logger.Log(Error.New("Failed to accept connection", err).Error())
				}
				return
			}
			go server.handleConnection(netConn)
		}
	}()
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
	server.tcpListener.Close()
	server.tcpListener = nil
	return nil
}

func (server *Server) IsStarted() bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.isStarted
}
