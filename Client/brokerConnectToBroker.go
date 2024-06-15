package Client

import (
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/Utilities"
)

func (client *Client) connectToBroker(resolution *Resolver.Resolution) (*brokerConnection, error) {
	netConn, err := client.tlsDial(resolution.Address, resolution.Certificate)
	if err != nil {
		return nil, Utilities.NewError("Error connecting to message broker server", err)
	}
	response, err := client.tcpExchange(netConn, Message.NewAsync("connect", client.name, ""))
	netConn.Close()
	if err != nil {
		return nil, Utilities.NewError("Error sending connection request", err)
	}
	if response.GetTopic() != "connected" {
		return nil, Utilities.NewError("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, resolution, client.logger), nil
}
