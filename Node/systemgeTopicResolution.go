package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (systemge *systemgeComponent) resolveBrokerForTopic(nodeName string, topic string) (*Config.TcpEndpoint, error) {
	netConn, err := Tcp.NewEndpoint(systemge.application.GetSystemgeComponentConfig().ResolverEndpoint)
	if err != nil {
		return nil, Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	responseMessage, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("resolve", nodeName, topic).Serialize(), systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		return nil, Error.New("failed to recieve response from resolver", err)
	}
	systemge.bytesSentCounter.Add(bytesSent)
	systemge.bytesReceivedCounter.Add(bytesReceived)
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

func (systemge *systemgeComponent) topicResolutionLifetimeTimeout(nodeName string, topic string, brokerConnection *brokerConnection) error {
	timer := time.NewTimer(time.Duration(systemge.application.GetSystemgeComponentConfig().TopicResolutionLifetimeMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		endpoint, err := systemge.resolveBrokerForTopic(nodeName, topic)
		if err != nil || endpoint.Address != brokerConnection.endpoint.Address {
			return Error.New("failed to re-resolve broker address for topic \""+topic+"\"", err)
		}
		return nil
	case <-brokerConnection.closeChannel:
		return Error.New("broker connection closed", nil)
	}
}
