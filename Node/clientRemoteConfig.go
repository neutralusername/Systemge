package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (client *Node) AddSyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("addSyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Node) RemoveSyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("removeSyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Node) AddAsyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("addAsyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Node) RemoveAsyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("removeAsyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Node) AddResolverTopicsRemotely(resolverAddress, nameIndication, tlsCertificate, brokerName string, topics ...string) error {
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	payload := brokerName
	for _, topic := range topics {
		payload += " " + topic
	}
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("addTopics", client.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (client *Node) RemoveResolverTopicsRemotely(resolverAddress, nameIndication, tlsCertificate string, topic ...string) error {
	if len(topic) == 0 {
		return Error.New("No topics provided", nil)
	}
	payload := ""
	for _, topic := range topic {
		payload += topic + " "
	}
	payload = payload[:len(payload)-1]
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("removeTopics", client.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (client *Node) AddKnownBrokerRemotely(resolverAddress, nameIndication, tlsCertificate string, resolution *Resolution.Resolution) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("addKnownBroker", client.GetName(), resolution.Marshal()), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Node) RemoveKnownBrokerRemotely(resolverAddress, nameIndication, tlsCertificate, brokerName string) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("removeKnownBroker", client.GetName(), brokerName), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with broker", err)
	}
	return nil
}
