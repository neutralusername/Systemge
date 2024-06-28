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

func (node *Node) AddResolverTopicRemotely(resolverAddress, nameIndication, tlsCertificate string, brokerResolution Resolution.Resolution, topic string) error {
	payload := brokerResolution.Marshal() + "|" + topic
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("addTopic", node.GetName(), payload), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (node *Node) RemoveResolverTopicRemotely(resolverAddress, nameIndication, tlsCertificate string, topic string) error {
	if len(topic) == 0 {
		return Error.New("No topics provided", nil)
	}
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("removeTopic", node.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error exchanging messages with topic resolution server", err)
	}
	return nil
}
