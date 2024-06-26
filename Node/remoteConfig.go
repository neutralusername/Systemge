package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (node *Node) AddSyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("addSyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveSyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("removeSyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) AddAsyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("addAsyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveAsyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("removeAsyncTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) AddResolverTopicsRemotely(resolverAddress, nameIndication, tlsCertificate, brokerName string, topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	payload := brokerName
	for _, topic := range topics {
		payload += " " + topic
	}
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("addTopics", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (node *Node) RemoveResolverTopicsRemotely(resolverAddress, nameIndication, tlsCertificate string, topic ...string) error {
	if len(topic) == 0 {
		return Error.New("No topics provided", nil)
	}
	payload := ""
	for _, topic := range topic {
		payload += topic + " "
	}
	payload = payload[:len(payload)-1]
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("removeTopics", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (node *Node) AddKnownBrokerRemotely(resolverAddress, nameIndication, tlsCertificate string, resolution *Resolution.Resolution) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("addKnownBroker", node.GetName(), resolution.Marshal()), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (node *Node) RemoveKnownBrokerRemotely(resolverAddress, nameIndication, tlsCertificate, brokerName string) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("removeKnownBroker", node.GetName(), brokerName), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}
