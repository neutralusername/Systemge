package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
)

func (broker *Broker) addResolverTopicRemotely(resolverConfigEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	netConn, err := resolverConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("Error dialing broker", err)
	}
	defer netConn.Close()
	payload := broker.config.Endpoint.Marshal() + "|" + topic
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("addTopic", broker.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with resolver", err)
	}
	return nil
}

func (broker *Broker) removeResolverTopicRemotely(resolverConfigEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	if len(topic) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := resolverConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("Error dialing broker", err)
	}
	defer netConn.Close()
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("removeTopic", broker.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with resolver", err)
	}
	return nil
}
