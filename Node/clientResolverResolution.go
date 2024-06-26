package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (client *Node) resolveBrokerForTopic(topic string) (*Resolution.Resolution, error) {
	netConn, err := Utilities.TlsDial(client.config.ResolverAddress, client.config.ResolverNameIndication, client.config.ResolverTLSCert)
	if err != nil {
		return nil, Error.New("Error dialing resolver", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("resolve", client.config.Name, topic), DEFAULT_TCP_TIMEOUT)
	netConn.Close()
	if err != nil {
		return nil, Error.New("Error resolving broker", err)
	}
	resolution := Resolution.Unmarshal(responseMessage.GetPayload())
	if resolution == nil {
		return nil, Error.New("Error unmarshalling broker", nil)
	}
	return resolution, nil
}

func (client *Node) GetTopicResolution(topic string) *brokerConnection {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	return client.topicResolutions[topic]
}

func (client *Node) addTopicResolution(topic string, serverConnection *brokerConnection) error {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	if client.topicResolutions[topic] != nil {
		return Error.New("Topic resolution already exists", nil)
	}
	err := serverConnection.addTopic(topic)
	if err != nil {
		return Error.New("Error adding topic to server connection", err)
	}
	client.topicResolutions[topic] = serverConnection
	return nil
}

// RemoveTopicResolution removes a topic resolution from the client
// Subscribed topics, i.e. topics with message handlers in the application, cannot be removed
func (client *Node) RemoveTopicResolution(topic string) error {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	serverConnection := client.topicResolutions[topic]
	if serverConnection == nil {
		return Error.New("Topic resolution does not exist", nil)
	}
	if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
		return Error.New("Cannot remove topics you are subscribed to", nil)
	}
	err := serverConnection.removeTopic(topic)
	if err != nil {
		return Error.New("Error removing topic from server connection", err)
	}
	delete(client.topicResolutions, topic)
	return nil
}
