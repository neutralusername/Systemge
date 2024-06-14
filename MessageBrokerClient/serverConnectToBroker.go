package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TCP"
	"net"
)

func (client *Client) connectToBroker(brokerAddress string) (*serverConnection, error) {
	netConn, err := net.Dial("tcp", brokerAddress)
	if err != nil {
		return nil, Error.New("Error connecting to message broker server", err)
	}
	err = TCP.Send(netConn, Message.NewAsync("connect", client.name, "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error sending connection request", err)
	}
	messageBytes, err := TCP.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error receiving connection response", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil {
		netConn.Close()
		return nil, Error.New("Invalid response \""+string(messageBytes)+"\"", nil)
	}
	if message.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("Invalid response topic \""+message.GetTopic()+"\"", nil)
	}
	serverConnection := newServerConnection(netConn, brokerAddress, client.logger)
	err = client.addServerConnection(serverConnection)
	if err != nil {
		serverConnection.close()
		return nil, Error.New("Error adding server connection", err)
	}
	return serverConnection, nil
}
