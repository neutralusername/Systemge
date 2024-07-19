package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"time"
)

func (node *Node) resolveBrokerForTopic(topic string) (*Config.TcpEndpoint, error) {
	netConn, err := Tcp.NewClient(node.GetSystemgeComponent().GetSystemgeComponentConfig().ResolverEndpoint)
	if err != nil {
		return nil, Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	responseMessage, err := Tcp.Exchange(netConn, Message.NewAsync("resolve", node.GetName(), topic), node.GetSystemgeComponent().GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return nil, Error.New("failed to recieve response from resolver", err)
	}
	if responseMessage.GetTopic() != "resolution" {
		return nil, Error.New("received error response from resolver \""+responseMessage.GetPayload()+"\"", nil)
	}
	endpoint := Config.UnmarshalTcpEndpoint(responseMessage.GetPayload())
	if endpoint == nil {
		return nil, Error.New("failed unmarshalling broker", nil)
	}
	return endpoint, nil
}

func (node *Node) getTopicResolution(topic string) *brokerConnection {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	return node.systemgeTopicResolutions[topic]
}

func (node *Node) addTopicResolution(topic string, brokerConnection *brokerConnection) error {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	if node.systemgeTopicResolutions[topic] != nil {
		return Error.New("topic endpoint already exists", nil)
	}
	err := brokerConnection.addTopicResolution(topic)
	if err != nil {
		return Error.New("failed adding topic to server connection", err)
	}
	node.systemgeTopicResolutions[topic] = brokerConnection
	go node.removeTopicResolutionTimeout(topic, brokerConnection)
	return nil
}

func (node *Node) removeTopicResolutionTimeout(topic string, brokerConnection *brokerConnection) {
	timer := time.NewTimer(time.Duration(node.GetSystemgeComponent().GetSystemgeComponentConfig().TopicResolutionLifetimeMs) * time.Millisecond)
	select {
	case <-timer.C:
		err := node.removeTopicResolution(topic)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed removing topic resolution", err).Error())
			}
		} else {
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Removed topic resolution for topic \""+topic+"\"", nil).Error())
			}
		}
	case <-node.stopChannel:
		timer.Stop()
	case <-brokerConnection.closeChannel:
		timer.Stop()
	}
}

func (node *Node) removeTopicResolution(topic string) error {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	brokerConnection := node.systemgeTopicResolutions[topic]
	if brokerConnection == nil {
		return Error.New("topic endpoint does not exist", nil)
	}
	err := brokerConnection.removeTopicResolution(topic)
	if err != nil {
		return Error.New("failed removing topic from server connection", err)
	}
	delete(node.systemgeTopicResolutions, topic)
	if len(brokerConnection.topicResolutions) == 0 && len(brokerConnection.subscribedTopics) == 0 {
		err = brokerConnection.close()
		if err != nil {
			return Error.New("failed closing broker connection", err)
		}
	}
	return nil
}
