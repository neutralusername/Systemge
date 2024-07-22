package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"time"
)

func (systemge *systemgeComponent) resolveBrokerForTopic(nodeName string, topic string) (*Config.TcpEndpoint, error) {
	netConn, err := Tcp.NewEndpoint(systemge.application.GetSystemgeComponentConfig().ResolverEndpoint)
	if err != nil {
		return nil, Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	responseMessage, err := Tcp.Exchange(netConn, Message.NewAsync("resolve", nodeName, topic), systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
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

func (systemge *systemgeComponent) getTopicResolution(topic string) *brokerConnection {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	return systemge.topicResolutions[topic]
}

func (systemge *systemgeComponent) addTopicResolution(topic string, brokerConnection *brokerConnection) error {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	if systemge.topicResolutions[topic] != nil {
		return Error.New("topic endpoint already exists", nil)
	}
	err := brokerConnection.addTopicResolution(topic)
	if err != nil {
		return Error.New("failed adding topic to server connection", err)
	}
	systemge.topicResolutions[topic] = brokerConnection
	return nil
}

func (systemge *systemgeComponent) removeTopicResolutionTimeout(topic string, brokerConnection *brokerConnection) error {
	timer := time.NewTimer(time.Duration(systemge.application.GetSystemgeComponentConfig().TopicResolutionLifetimeMs) * time.Millisecond)
	select {
	case <-timer.C:
		err := systemge.removeTopicResolution(topic)
		if err != nil {
			return Error.New("failed removing topic resolution", err)
		}
		return nil
	case <-brokerConnection.closeChannel:
		timer.Stop()
		return Error.New("broker connection closed", nil)
	}
}

func (systemge *systemgeComponent) removeTopicResolution(topic string) error {
	if systemge == nil {
		return Error.New("systemge not initialized", nil)
	}
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	brokerConnection := systemge.topicResolutions[topic]
	if brokerConnection == nil {
		return Error.New("topic endpoint does not exist", nil)
	}
	err := brokerConnection.removeTopicResolution(topic)
	if err != nil {
		return Error.New("failed removing topic from server connection", err)
	}
	delete(systemge.topicResolutions, topic)
	if len(brokerConnection.topicResolutions) == 0 && len(brokerConnection.subscribedTopics) == 0 {
		err = brokerConnection.close()
		if err != nil {
			return Error.New("failed closing broker connection", err)
		}
	}
	return nil
}
