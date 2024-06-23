package Client

import (
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (client *Client) resolveBrokerForTopic(resolverResolution *Resolution.Resolution, topic string) (*Resolution.Resolution, error) {
	netConn, err := Utilities.TlsDial(resolverResolution.GetAddress(), resolverResolution.GetServerNameIndication(), resolverResolution.GetTlsCertificate())
	if err != nil {
		return nil, Utilities.NewError("Error dialing resolver", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("resolve", client.name, topic), DEFAULT_TCP_TIMEOUT)
	netConn.Close()
	if err != nil {
		return nil, Utilities.NewError("Error resolving broker", err)
	}
	resolution := Resolution.Unmarshal(responseMessage.GetPayload())
	if resolution == nil {
		return nil, Utilities.NewError("Error unmarshalling broker", nil)
	}
	return resolution, nil
}

func (client *Client) getTopicResolution(topic string) *brokerConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.topicResolutions[topic]
}

func (client *Client) addTopicResolution(topic string, serverConnection *brokerConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.topicResolutions[topic] != nil {
		return Utilities.NewError("Topic resolution already exists", nil)
	}
	err := serverConnection.addTopic(topic)
	if err != nil {
		return Utilities.NewError("Error adding topic to server connection", err)
	}
	client.topicResolutions[topic] = serverConnection
	return nil
}

func (client *Client) RemoveTopicResolution(topic string) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	serverConnection := client.topicResolutions[topic]
	if serverConnection == nil {
		return Utilities.NewError("Topic resolution does not exist", nil)
	}
	if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
		return Utilities.NewError("Cannot remove topics you are subscribed to", nil)
	}
	err := serverConnection.removeTopic(topic)
	if err != nil {
		return Utilities.NewError("Error removing topic from server connection", err)
	}
	delete(client.topicResolutions, topic)
	return nil
}
