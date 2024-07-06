package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
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
	return nil
}
