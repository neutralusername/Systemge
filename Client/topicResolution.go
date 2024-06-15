package Client

import (
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/Utilities"
	"net"
)

func (client *Client) resolveBrokerForTopic(topic string) (*Resolver.Resolution, error) {
	netConn, err := net.Dial("tcp", client.resolverAddress)
	if err != nil {
		return nil, Utilities.NewError("Error connecting to topic resolution server", err)
	}
	err = Utilities.Send(netConn, Message.NewAsync("resolve", client.name, topic).Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error sending topic resolution request", err)
	}
	messageBytes, err := Utilities.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error receiving topic resolution response", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolution" {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response \""+string(messageBytes)+"\"", nil)
	}
	resolution := Resolver.UnmarshalResolution(message.GetPayload())
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
