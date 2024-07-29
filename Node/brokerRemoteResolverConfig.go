package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (broker *brokerComponent) addResolverTopicsRemotely(resolverConfigEndpoint *Config.TcpEndpoint, nodeName string, topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(resolverConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	payload := Helpers.JsonMarshal(broker.application.GetBrokerComponentConfig().Endpoint)
	for _, topic := range topics {
		payload += "|" + topic
	}
	response, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("addTopics", nodeName, payload).Serialize(), broker.application.GetBrokerComponentConfig().TcpTimeoutMs, broker.application.GetBrokerComponentConfig().IncomingMessageByteLimit)
	if err != nil {
		return Error.New("failed exchanging messages with resolver", err)
	}
	broker.bytesSentCounter.Add(bytesSent)
	broker.bytesReceivedCounter.Add(bytesReceived)
	if response.GetTopic() == "error" {
		return Error.New("resolver config request failed", Error.New(response.GetPayload(), nil))
	}
	return nil
}

func (broker *brokerComponent) removeResolverTopicsRemotely(resolverConfigEndpoint *Config.TcpEndpoint, nodeName string, topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(resolverConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	payload := Helpers.JsonMarshal(broker.application.GetBrokerComponentConfig().Endpoint)
	for _, topic := range topics {
		payload += "|" + topic
	}
	response, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("removeTopics", nodeName, payload).Serialize(), broker.application.GetBrokerComponentConfig().TcpTimeoutMs, broker.application.GetBrokerComponentConfig().IncomingMessageByteLimit)
	if err != nil {
		return Error.New("failed exchanging messages with resolver", err)
	}
	broker.bytesSentCounter.Add(bytesSent)
	broker.bytesReceivedCounter.Add(bytesReceived)
	if response.GetTopic() == "error" {
		return Error.New("resolver config request failed", Error.New(response.GetPayload(), nil))
	}
	return nil
}
