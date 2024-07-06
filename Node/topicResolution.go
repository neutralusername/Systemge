package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
	"time"
)

func (node *Node) resolveBrokerForTopic(topic string) (*TcpEndpoint.TcpEndpoint, error) {
	netConn, err := node.config.ResolverEndpoint.TlsDial()
	if err != nil {
		return nil, Error.New("Error dialing resolver", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("resolve", node.config.Name, topic), DEFAULT_TCP_TIMEOUT)
	netConn.Close()
	if err != nil {
		return nil, Error.New("Error resolving broker", err)
	}
	endpoint := TcpEndpoint.Unmarshal(responseMessage.GetPayload())
	if endpoint == nil {
		return nil, Error.New("Error unmarshalling broker", nil)
	}
	return endpoint, nil
}

func (node *Node) getTopicBrokerConnection(topic string) *brokerConnection {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	return node.topicBrokerConnections[topic]
}

func (node *Node) addTopicBrokerConnection(topic string, brokerConnection *brokerConnection) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.topicBrokerConnections[topic] != nil {
		return Error.New("Topic endpoint already exists", nil)
	}
	err := brokerConnection.addTopic(topic)
	if err != nil {
		return Error.New("Error adding topic to server connection", err)
	}
	node.topicBrokerConnections[topic] = brokerConnection
	go node.removeTopicBrokerConnectionTimeout(topic)
	return nil
}

func (node *Node) removeTopicBrokerConnectionTimeout(topic string) {
	timer := time.NewTimer(time.Duration(node.config.TopicResolutionLifetimeMs) * time.Millisecond)
	select {
	case <-timer.C:
		err := node.removeTopicBrokerConnection(topic)
		if err != nil {
			node.logger.Log(Error.New("Error removing topic broker connection", err).Error())
		}
	case <-node.stopChannel:
		timer.Stop()
	}
}

func (node *Node) removeTopicBrokerConnection(topic string) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	brokerConnection := node.topicBrokerConnections[topic]
	if brokerConnection == nil {
		return Error.New("Topic endpoint does not exist", nil)
	}
	err := brokerConnection.removeTopic(topic)
	if err != nil {
		return Error.New("Error removing topic from server connection", err)
	}
	delete(node.topicBrokerConnections, topic)
	return nil
}
