package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
)

func (broker *Broker) addResolverTopicsRemotely(topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := broker.config.ResolverConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	payload := broker.config.Endpoint.Marshal()
	for _, topic := range topics {
		payload += "|" + topic
	}
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("addTopics", broker.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed exchanging messages with resolver", err)
	}
	return nil
}

func (broker *Broker) removeResolverTopicsRemotely(topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	netConn, err := broker.config.ResolverConfigEndpoint.TlsDial()
	if err != nil {
		return Error.New("failed dialing broker", err)
	}
	defer netConn.Close()
	payload := ""
	for _, topic := range topics {
		payload += topic + "|"
	}
	payload = payload[:len(payload)-1]
	_, err = Utilities.TcpExchange(netConn, Message.NewAsync("removeTopics", broker.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed exchanging messages with resolver", err)
	}
	return nil
}
