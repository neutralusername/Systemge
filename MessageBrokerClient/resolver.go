package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TCP"
	"net"
)

func (client *Client) resolveBrokerAddressForTopic(topic string) (string, error) {
	netConn, err := net.Dial("tcp", client.topicResolutionServerAddress)
	if err != nil {
		return "", Error.New("Error connecting to topic resolution server", err)
	}
	err = TCP.Send(netConn, []byte(Message.NewAsync("resolve", client.name, topic).Serialize()), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return "", Error.New("Error sending topic resolution request", err)
	}
	messageBytes, err := TCP.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return "", Error.New("Error receiving topic resolution response", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolution" {
		netConn.Close()
		return "", Error.New("Invalid response \""+string(messageBytes)+"\"", nil)
	}
	return message.GetPayload(), nil
}
