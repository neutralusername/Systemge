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
		return nil, Error.New("failed dialing resolver", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("resolve", node.config.Name, topic), DEFAULT_TCP_TIMEOUT)
	netConn.Close()
	if err != nil {
		return nil, Error.New("failed resolving broker", err)
	}
	endpoint := TcpEndpoint.Unmarshal(responseMessage.GetPayload())
	if endpoint == nil {
		return nil, Error.New("failed unmarshalling broker", nil)
	}
	return endpoint, nil
}

func (node *Node) getTopicResolution(topic string) *brokerConnection {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	return node.topicResolutions[topic]
}

func (node *Node) addTopicResolution(topic string, brokerConnection *brokerConnection) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.topicResolutions[topic] != nil {
		return Error.New("topic endpoint already exists", nil)
	}
	err := brokerConnection.addTopicResolution(topic)
	if err != nil {
		return Error.New("failed adding topic to server connection", err)
	}
	node.topicResolutions[topic] = brokerConnection
	go node.removeTopicResolutionTimeout(topic, brokerConnection)
	return nil
}

func (node *Node) removeTopicResolutionTimeout(topic string, brokerConnection *brokerConnection) {
	timer := time.NewTimer(time.Duration(node.config.TopicResolutionLifetimeMs) * time.Millisecond)
	select {
	case <-timer.C:
		err := node.removeTopicResolution(topic)
		if err != nil {
			node.config.Logger.Warning(Error.New("Failed removing topic resolution", err).Error())
		} else {
			node.config.Logger.Info(Error.New("Removed topic resolution for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
		}
	case <-node.stopChannel:
		timer.Stop()
	case <-brokerConnection.closeChannel:
		timer.Stop()
	}
}

func (node *Node) removeTopicResolution(topic string) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	brokerConnection := node.topicResolutions[topic]
	if brokerConnection == nil {
		return Error.New("topic endpoint does not exist", nil)
	}
	err := brokerConnection.removeTopicResolution(topic)
	if err != nil {
		return Error.New("failed removing topic from server connection", err)
	}
	delete(node.topicResolutions, topic)
	if len(brokerConnection.topicResolutions) == 0 && len(brokerConnection.subscribedTopics) == 0 {
		err = brokerConnection.close()
		if err != nil {
			return Error.New("failed closing broker connection", err)
		}
	}
	return nil
}
