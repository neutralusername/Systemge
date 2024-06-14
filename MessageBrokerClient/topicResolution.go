package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/ResolverServer"
	"Systemge/TCP"
	"net"
)

func (client *Client) resolveBrokerForTopic(topic string) (*ResolverServer.Broker, error) {
	netConn, err := net.Dial("tcp", client.topicResolutionServerAddress)
	if err != nil {
		return nil, Error.New("Error connecting to topic resolution server", err)
	}
	err = TCP.Send(netConn, Message.NewAsync("resolve", client.name, topic).Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error sending topic resolution request", err)
	}
	messageBytes, err := TCP.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error receiving topic resolution response", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolution" {
		netConn.Close()
		return nil, Error.New("Invalid response \""+string(messageBytes)+"\"", nil)
	}
	broker := ResolverServer.UnmarshalBroker(message.GetPayload())
	if broker == nil {
		netConn.Close()
		return nil, Error.New("Error unmarshalling broker", nil)
	}
	return broker, nil
}

func (client *Client) addTopicResolution(topic string, serverConnection *serverConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.topicResolutions[topic] != nil {
		return Error.New("Topic resolution already exists", nil)
	}
	client.topicResolutions[topic] = serverConnection
	serverConnection.addTopic(topic)
	return nil
}
func (serverConnection *serverConnection) addTopic(topic string) {
	serverConnection.mapOperationMutex.Lock()
	serverConnection.topics[topic] = true
	serverConnection.mapOperationMutex.Unlock()
}

func (client *Client) getTopicResolution(topic string) *serverConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.topicResolutions[topic]
}
