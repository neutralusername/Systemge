package Client

import (
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/Utilities"
)

func (client *Client) AddSyncTopicRemotely(address, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("addSyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) RemoveSyncTopicRemotely(address, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("removeSyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) AddAsyncTopicRemotely(address, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("addAsyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) RemoveAsyncTopicRemotely(address, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("removeAsyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) RegisterTopicRemotely(address, nameIndication, tlsCertificate, brokerName, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("registerTopics", client.GetName(), brokerName+" "+topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (client *Client) UnregisterTopicRemotely(address, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("unregisterTopics", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (client *Client) RegisterBrokerRemotely(address, nameIndication, tlsCertificate string, resolution *Resolver.Resolution) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("registerBroker", client.GetName(), resolution.Marshal()), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) UnregisterBrokerRemotely(address, nameIndication, tlsCertificate, brokerName string) error {
	_, err := Utilities.TcpOneTimeExchange(address, nameIndication, tlsCertificate, Message.NewAsync("unregisterBroker", client.GetName(), brokerName), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}
