package Node

import (
	"net"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
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

func (brokerConnection *brokerConnection) send(tcpTimeoutMs uint64, messageBytes []byte) (uint64, error) {
	brokerConnection.sendMutex.Lock()
	defer brokerConnection.sendMutex.Unlock()
	if brokerConnection.netConn == nil {
		return 0, Error.New("Connection is closed", nil)
	}
	bytesSent, err := Tcp.Send(brokerConnection.netConn, messageBytes, tcpTimeoutMs)
	if err != nil {
		return 0, Error.New("Failed sending message", err)
	}
	return bytesSent, nil
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

func (brokerConnection *brokerConnection) removeTopicResolution(topic string) {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	delete(brokerConnection.topicResolutions, topic)
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

func (brokerConnection *brokerConnection) closeIfNoTopics() (closed bool) {
	brokerConnection.mutex.Lock()
	defer brokerConnection.mutex.Unlock()
	subscribedTopicsCount := len(brokerConnection.subscribedTopics)
	topicResolutionsCount := len(brokerConnection.topicResolutions)
	if subscribedTopicsCount == 0 && topicResolutionsCount == 0 {
		return brokerConnection.closeNetConn(false)
	}
	return false
}
func (brokerConnection *brokerConnection) closeNetConn(lock bool) (closed bool) {
	if lock {
		brokerConnection.mutex.Lock()
		defer brokerConnection.mutex.Unlock()
	}
	if brokerConnection.netConn == nil {
		return false
	}
	brokerConnection.netConn.Close()
	brokerConnection.netConn = nil
	return true
}

func (systemge *systemgeComponent) cleanUpDisconnectedBrokerConnection(brokerConnection *brokerConnection) (previouslySubscribedTopics []string) {
	systemge.brokerConnectionsMutex.Lock()
	brokerConnection.mutex.Lock()
	defer func() {
		brokerConnection.mutex.Unlock()
		systemge.brokerConnectionsMutex.Unlock()
	}()
	delete(systemge.brokerConnections, brokerConnection.endpoint.Address)
	brokerConnection.topicResolutions = make(map[string]bool)
	previouslySubscribedTopics = make([]string, 0)
	for topic := range brokerConnection.subscribedTopics {
		previouslySubscribedTopics = append(previouslySubscribedTopics, topic)
		delete(brokerConnection.subscribedTopics, topic)
	}
	return previouslySubscribedTopics
}
