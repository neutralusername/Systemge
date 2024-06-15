package Client

import (
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/Utilities"
)

func (client *Client) connectToBroker(resolution *Resolver.Resolution) (*brokerConnection, error) {
	netConn, err := Utilities.TlsDial(resolution.Address, resolution.Certificate)
	if err != nil {
		return nil, Utilities.NewError("Error connecting to message broker server", err)
	}
	responseBytes, err := Utilities.TcpExchange(netConn, Message.NewAsync("connect", client.name, "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error sending connection request", err)
	}
	responseMessage := Message.Deserialize(responseBytes)
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, resolution, client.logger), nil
}
