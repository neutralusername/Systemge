package Broker

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (broker *Broker) addResolverTopicsRemotely(topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(broker.config.ResolverConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	payload := broker.config.Endpoint.Marshal()
	for _, topic := range topics {
		payload += "|" + topic
	}
	response, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("addTopics", broker.node.GetName(), payload).Serialize(), broker.config.TcpTimeoutMs, broker.config.IncomingMessageByteLimit)
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

func (broker *Broker) removeResolverTopicsRemotely(topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := Tcp.NewEndpoint(broker.config.ResolverConfigEndpoint)
	if err != nil {
		return Error.New("failed dialing resolver", err)
	}
	defer netConn.Close()
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	payload = payload[:len(payload)-1]
	response, bytesSent, bytesReceived, err := Tcp.Exchange(netConn, Message.NewAsync("removeTopics", broker.node.GetName(), payload).Serialize(), broker.config.TcpTimeoutMs, broker.config.IncomingMessageByteLimit)
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
