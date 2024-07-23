package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Tcp"
	"net"
	"sync"
)

type brokerConnection struct {
	netConn  net.Conn
	endpoint *Config.TcpEndpoint

	topicResolutions map[string]bool
	subscribedTopics map[string]bool

	closeChannel chan bool

	mutex        sync.Mutex
	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func newBrokerConnection(netConn net.Conn, tcpEndpoint *Config.TcpEndpoint) *brokerConnection {
	return &brokerConnection{
		netConn:  netConn,
		endpoint: tcpEndpoint,

		closeChannel: make(chan bool),

		topicResolutions: make(map[string]bool),
		subscribedTopics: make(map[string]bool),
	}
}

func (brokerConnection *brokerConnection) send(tcpTimeoutMs uint64, messageBytes []byte) error {
	brokerConnection.sendMutex.Lock()
	defer brokerConnection.sendMutex.Unlock()
	if brokerConnection.netConn == nil {
		return Error.New("Connection is closed", nil)
	}
	err := Tcp.Send(brokerConnection.netConn, messageBytes, tcpTimeoutMs)
	if err != nil {
		return Error.New("Failed sending message", err)
	}
	return nil
}

func (brokerConnection *brokerConnection) close() error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if brokerConnection.netConn == nil {
		return Error.New("Connection is already closed", nil)
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
		return Error.New("Topic already exists", nil)
	}
	brokerConnection.topicResolutions[topic] = true
	return nil
}

func (brokerConnection *brokerConnection) removeTopicResolution(topic string) error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if !brokerConnection.topicResolutions[topic] {
		return Error.New("Topic does not exist", nil)
	}
	delete(brokerConnection.topicResolutions, topic)
	return nil
}

func (brokerConnection *brokerConnection) addSubscribedTopic(topic string) error {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	if brokerConnection.subscribedTopics[topic] {
		return Error.New("Topic already exists", nil)
	}
	brokerConnection.subscribedTopics[topic] = true
	return nil
}
