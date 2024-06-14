package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/ResolverServer"
	"Systemge/TCP"
	"Systemge/Utilities"
	"net"
	"sync"
)

type serverConnection struct {
	netConn net.Conn
	broker  *ResolverServer.Broker
	logger  *Utilities.Logger

	topics            map[string]bool
	mapOperationMutex sync.Mutex

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func newServerConnection(netConn net.Conn, broker *ResolverServer.Broker, logger *Utilities.Logger) *serverConnection {
	return &serverConnection{
		netConn: netConn,
		broker:  broker,
		logger:  logger,

		topics: make(map[string]bool),
	}
}

func (serverConnection *serverConnection) send(message *Message.Message) error {
	if serverConnection == nil {
		return Error.New("Server connection is nil", nil)
	}
	serverConnection.sendMutex.Lock()
	defer serverConnection.sendMutex.Unlock()
	if serverConnection.netConn == nil {
		return Error.New("Connection is closed", nil)
	}
	err := TCP.Send(serverConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error sending message", err)
	}
	return nil
}

func (serverConnection *serverConnection) receive() ([]byte, error) {
	if serverConnection == nil {
		return nil, Error.New("Server connection is nil", nil)
	}
	serverConnection.receiveMutex.Lock()
	defer serverConnection.receiveMutex.Unlock()
	if serverConnection.netConn == nil {
		return nil, Error.New("Connection is closed", nil)
	}
	messageBytes, err := TCP.Receive(serverConnection.netConn, 0)
	if err != nil {
		return nil, Error.New("Error receiving message", err)
	}
	return messageBytes, nil
}

func (serverConnection *serverConnection) close() error {
	if serverConnection == nil {
		return Error.New("Server connection is nil", nil)
	}
	if serverConnection.netConn == nil {
		return Error.New("Connection is already closed", nil)
	}
	serverConnection.netConn.Close()
	serverConnection.netConn = nil
	return nil
}

func (client *Client) attemptToReconnect(serverConnection *serverConnection) {
	client.mapOperationMutex.Lock()
	serverConnection.mapOperationMutex.Lock()
	delete(client.serverConnections, serverConnection.broker.Name)
	topicsToReconnect := make([]string, 0)
	for topic := range serverConnection.topics {
		delete(client.topicResolutions, topic)
		if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
			topicsToReconnect = append(topicsToReconnect, topic)
		}
	}
	serverConnection.topics = make(map[string]bool)
	serverConnection.mapOperationMutex.Unlock()
	client.mapOperationMutex.Unlock()

	for _, topic := range topicsToReconnect {
		newServerConnection, err := client.getServerConnectionForTopic(topic)
		if err != nil {
			panic(err)
		}
		err = client.subscribeTopic(newServerConnection, topic)
		if err != nil {
			panic(err)
		}
	}
}
