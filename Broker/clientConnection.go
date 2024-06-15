package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"sync"
	"time"
)

type clientConnection struct {
	name         string
	netConn      net.Conn
	messageQueue chan *Message.Message
	watchdog     *time.Timer

	subscribedTopics   map[string]bool
	openSyncRequests   map[string]*Message.Message
	deliverImmediately bool

	mutex        sync.Mutex
	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func newClientConnection(name string, netConn net.Conn) *clientConnection {
	return &clientConnection{
		name:               name,
		netConn:            netConn,
		messageQueue:       make(chan *Message.Message, CLIENT_MESSAGE_QUEUE_SIZE),
		watchdog:           nil,
		openSyncRequests:   map[string]*Message.Message{},
		subscribedTopics:   map[string]bool{},
		deliverImmediately: DELIVER_IMMEDIATELY_DEFAULT,
	}
}

func (clientConnection *clientConnection) send(message *Message.Message) error {
	clientConnection.sendMutex.Lock()
	defer clientConnection.sendMutex.Unlock()
	return Utilities.TcpSend(clientConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
}

func (clientConnection *clientConnection) receive() ([]byte, error) {
	clientConnection.receiveMutex.Lock()
	defer clientConnection.receiveMutex.Unlock()
	messageBytes, err := Utilities.TcpReceive(clientConnection.netConn, 0)
	if err != nil {
		return nil, err
	}
	return messageBytes, nil
}

func (clientConnection *clientConnection) resetWatchdog() error {
	clientConnection.mutex.Lock()
	defer clientConnection.mutex.Unlock()
	if clientConnection.watchdog == nil {
		return Utilities.NewError("Watchdog is not set for client \""+clientConnection.name+"\"", nil)
	}
	clientConnection.watchdog.Reset((WATCHDOG_TIMEOUT))
	return nil
}

func (clientConnection *clientConnection) disconnect() error {
	clientConnection.mutex.Lock()
	defer clientConnection.mutex.Unlock()
	if clientConnection.watchdog == nil {
		return Utilities.NewError("Watchdog is not set for client \""+clientConnection.name+"\"", nil)
	}
	clientConnection.watchdog.Reset(0)
	clientConnection.watchdog = nil
	clientConnection.netConn.Close()
	return nil
}

func (server *Server) addClient(clientConnection *clientConnection) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.clientConnections[clientConnection.name] != nil {
		return Utilities.NewError("client with name \""+clientConnection.name+"\" already exists", nil)
	}
	server.clientConnections[clientConnection.name] = clientConnection
	clientConnection.watchdog = time.AfterFunc(WATCHDOG_TIMEOUT, func() {
		server.logger.Log("Client \"" + clientConnection.name + "\" has timed out")
		server.removeClient(clientConnection)
	})
	return nil
}

func (server *Server) removeClient(clientConnection *clientConnection) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if server.clientConnections[clientConnection.name] == nil {
		return Utilities.NewError("Subscriber with name \""+clientConnection.name+"\" does not exist", nil)
	}
	clientConnection.watchdog = nil
	clientConnection.netConn.Close()
	for messageType := range clientConnection.subscribedTopics {
		delete(server.clientSubscriptions[messageType], clientConnection.name)
	}
	for _, message := range clientConnection.openSyncRequests {
		delete(server.openSyncRequests, message.GetSyncRequestToken())
	}
	delete(server.clientConnections, clientConnection.name)
	return nil
}
