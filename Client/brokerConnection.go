package Client

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/TCP"
	"Systemge/Utilities"
	"net"
	"sync"
)

type brokerConnection struct {
	netConn    net.Conn
	resolution *Resolver.Resolution
	logger     *Utilities.Logger

	topics            map[string]bool
	mapOperationMutex sync.Mutex

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func newBrokerConnection(netConn net.Conn, resolution *Resolver.Resolution, logger *Utilities.Logger) *brokerConnection {
	return &brokerConnection{
		netConn:    netConn,
		resolution: resolution,
		logger:     logger,

		topics: make(map[string]bool),
	}
}

func (brokerConnection *brokerConnection) send(message *Message.Message) error {
	if brokerConnection == nil {
		return Error.New("Server connection is nil", nil)
	}
	brokerConnection.sendMutex.Lock()
	defer brokerConnection.sendMutex.Unlock()
	if brokerConnection.netConn == nil {
		return Error.New("Connection is closed", nil)
	}
	err := TCP.Send(brokerConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error sending message", err)
	}
	return nil
}

func (brokerConnection *brokerConnection) receive() ([]byte, error) {
	if brokerConnection == nil {
		return nil, Error.New("Server connection is nil", nil)
	}
	brokerConnection.receiveMutex.Lock()
	defer brokerConnection.receiveMutex.Unlock()
	if brokerConnection.netConn == nil {
		return nil, Error.New("Connection is closed", nil)
	}
	messageBytes, err := TCP.Receive(brokerConnection.netConn, 0)
	if err != nil {
		return nil, Error.New("Error receiving message", err)
	}
	return messageBytes, nil
}

func (brokerConnection *brokerConnection) close() error {
	if brokerConnection == nil {
		return Error.New("Server connection is nil", nil)
	}
	if brokerConnection.netConn == nil {
		return Error.New("Connection is already closed", nil)
	}
	brokerConnection.netConn.Close()
	brokerConnection.netConn = nil
	return nil
}

func (brokerConnection *brokerConnection) addTopic(topic string) error {
	brokerConnection.mapOperationMutex.Lock()
	defer brokerConnection.mapOperationMutex.Unlock()
	if brokerConnection.topics[topic] {
		return Error.New("Topic already exists", nil)
	}
	brokerConnection.topics[topic] = true
	return nil
}
