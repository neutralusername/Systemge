package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (node *Node) resolveBrokerForTopic(topic string) (*Resolution.Resolution, error) {
	netConn, err := Utilities.TlsDial(node.application.GetApplicationConfig().ResolverAddress, node.application.GetApplicationConfig().ResolverNameIndication, node.application.GetApplicationConfig().ResolverTLSCert)
	if err != nil {
		return nil, Error.New("Error dialing resolver", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("resolve", node.config.Name, topic), DEFAULT_TCP_TIMEOUT)
	netConn.Close()
	if err != nil {
		return nil, Error.New("Error resolving broker", err)
	}
	resolution := Resolution.Unmarshal(responseMessage.GetPayload())
	if resolution == nil {
		return nil, Error.New("Error unmarshalling broker", nil)
	}
	return resolution, nil
}

func (node *Node) GetTopicResolution(topic string) *brokerConnection {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	return node.topicResolutions[topic]
}

func (node *Node) addTopicResolution(topic string, serverConnection *brokerConnection) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.topicResolutions[topic] != nil {
		return Error.New("Topic resolution already exists", nil)
	}
	err := serverConnection.addTopic(topic)
	if err != nil {
		return Error.New("Error adding topic to server connection", err)
	}
	node.topicResolutions[topic] = serverConnection
	return nil
}

// RemoveTopicResolution removes a topic resolution from the node
// Subscribed topics, i.e. topics with message handlers in the application, cannot be removed
func (node *Node) RemoveTopicResolution(topic string) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	serverConnection := node.topicResolutions[topic]
	if serverConnection == nil {
		return Error.New("Topic resolution does not exist", nil)
	}
	if node.application.GetAsyncMessageHandlers()[topic] != nil || node.application.GetSyncMessageHandlers()[topic] != nil {
		return Error.New("Cannot remove topics you are subscribed to", nil)
	}
	err := serverConnection.removeTopic(topic)
	if err != nil {
		return Error.New("Error removing topic from server connection", err)
	}
	delete(node.topicResolutions, topic)
	return nil
}
