package Broker

import (
	"Systemge/Error"
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
	deliverImmediately bool

	watchdogMutex sync.Mutex
	sendMutex     sync.Mutex
	receiveMutex  sync.Mutex

	stopChannel chan bool
}

func newClientConnection(name string, netConn net.Conn) *clientConnection {
	return &clientConnection{
		name:               name,
		netConn:            netConn,
		messageQueue:       make(chan *Message.Message, CLIENT_MESSAGE_QUEUE_SIZE),
		watchdog:           nil,
		subscribedTopics:   map[string]bool{},
		deliverImmediately: DELIVER_IMMEDIATELY_DEFAULT,
		stopChannel:        make(chan bool),
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
	clientConnection.watchdogMutex.Lock()
	defer clientConnection.watchdogMutex.Unlock()
	if clientConnection.watchdog == nil {
		return Error.New("Watchdog is not set for client \""+clientConnection.name+"\"", nil)
	}
	clientConnection.watchdog.Reset((WATCHDOG_TIMEOUT))
	return nil
}

func (clientConnection *clientConnection) disconnect() error {
	clientConnection.watchdogMutex.Lock()
	if clientConnection.watchdog == nil {
		return Error.New("Watchdog is not set for client \""+clientConnection.name+"\"", nil)
	}
	clientConnection.watchdog.Reset(0)
	<-clientConnection.stopChannel
	clientConnection.watchdogMutex.Unlock()
	return nil
}

func (server *Server) addClientConnection(clientConnection *clientConnection) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.clientConnections[clientConnection.name] != nil {
		return Error.New("client with name \""+clientConnection.name+"\" already exists", nil)
	}
	server.clientConnections[clientConnection.name] = clientConnection
	clientConnection.watchdog = time.AfterFunc(WATCHDOG_TIMEOUT, func() {
		watchdog := clientConnection.watchdog
		clientConnection.watchdog = nil
		watchdog.Stop()
		clientConnection.netConn.Close()
		err := server.removeClientConnection(clientConnection)
		if err != nil {
			server.logger.Log(Error.New("Error removing client \""+clientConnection.name+"\"", err).Error())
		}
		close(clientConnection.stopChannel)
	})
	return nil
}

func (server *Server) removeClientConnection(clientConnection *clientConnection) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.clientConnections[clientConnection.name] == nil {
		return Error.New("Subscriber with name \""+clientConnection.name+"\" does not exist", nil)
	}
	for messageType := range clientConnection.subscribedTopics {
		delete(server.clientSubscriptions[messageType], clientConnection.name)
	}
	delete(server.clientConnections, clientConnection.name)
	return nil
}

func (server *Server) disconnectAllClientConnections() {
	clientsToDisconnect := make([]*clientConnection, 0)
	server.operationMutex.Lock()
	for _, clientConnection := range server.clientConnections {
		clientsToDisconnect = append(clientsToDisconnect, clientConnection)
	}
	server.operationMutex.Unlock()
	for _, clientConnection := range clientsToDisconnect {
		clientConnection.disconnect()
	}
}
