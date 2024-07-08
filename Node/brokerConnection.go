package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
	"net"
	"sync"
)

type brokerConnection struct {
	netConn  net.Conn
	endpoint *TcpEndpoint.TcpEndpoint
	logger   *Utilities.Logger

	topicResolutions map[string]bool
	subscribedTopics map[string]bool

	closeChannel chan bool

	mutex        sync.Mutex
	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func newBrokerConnection(netConn net.Conn, tcpEndpoint *TcpEndpoint.TcpEndpoint, logger *Utilities.Logger) *brokerConnection {
	return &brokerConnection{
		netConn:  netConn,
		endpoint: tcpEndpoint,
		logger:   logger,

		closeChannel: make(chan bool),

		topicResolutions: make(map[string]bool),
		subscribedTopics: make(map[string]bool),
	}
}

func (brokerConnection *brokerConnection) send(message *Message.Message) error {
	brokerConnection.sendMutex.Lock()
	defer brokerConnection.sendMutex.Unlock()
	if brokerConnection.netConn == nil {
		return Error.New("connection is closed", nil)
	}
	err := Utilities.TcpSend(brokerConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed sending message", err)
	}
	return nil
}

func (brokerConnection *brokerConnection) receive() (*Message.Message, error) {
	brokerConnection.receiveMutex.Lock()
	defer brokerConnection.receiveMutex.Unlock()
	if brokerConnection.netConn == nil {
		return nil, Error.New("connection is closed", nil)
	}
	messageBytes, err := Utilities.TcpReceive(brokerConnection.netConn, 0)
	if err != nil {
		return nil, Error.New("failed receiving message", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil {
		return nil, Error.New("failed to deserialize message \""+string(messageBytes)+"\"", nil)
	}
	return message, nil
}

func (brokerConnection *brokerConnection) close() error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if brokerConnection.netConn == nil {
		return Error.New("connection is already closed", nil)
	}
	brokerConnection.netConn.Close()
	brokerConnection.netConn = nil
	close(brokerConnection.closeChannel)
	return nil
}

func (brokerConnection *brokerConnection) addTopicResolution(topic string) error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if brokerConnection.topicResolutions[topic] {
		return Error.New("topic already exists", nil)
	}
	brokerConnection.topicResolutions[topic] = true
	return nil
}

func (brokerConnection *brokerConnection) removeTopicResolution(topic string) error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if !brokerConnection.topicResolutions[topic] {
		return Error.New("topic does not exist", nil)
	}
	delete(brokerConnection.topicResolutions, topic)
	return nil
}

func (brokerConnection *brokerConnection) addSubscribedTopic(topic string) error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if brokerConnection.subscribedTopics[topic] {
		return Error.New("topic already exists", nil)
	}
	brokerConnection.subscribedTopics[topic] = true
	return nil
}
