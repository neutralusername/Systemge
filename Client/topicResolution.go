package Client

import (
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/Utilities"
)

func (client *Client) resolveBrokerForTopic(topic string) (*Resolver.Resolution, error) {
	netConn, err := client.tcpDial(client.resolverAddress)
	if err != nil {
		return nil, Utilities.NewError("Error dialing resolver", err)
	}
	response, err := client.tcpExchange(netConn, Message.NewAsync("resolve", client.name, topic))
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error resolving broker", err)
	}
	resolution := Resolver.UnmarshalResolution(response.GetPayload())
	if resolution == nil {
		netConn.Close()
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
