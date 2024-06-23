package Client

import (
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (client *Client) connectToBroker(resolution *Resolution.Resolution) (*brokerConnection, error) {
	netConn, err := Utilities.TlsDial(resolution.Address, resolution.ServerNameIndication, resolution.Certificate)
	if err != nil {
		return nil, Utilities.NewError("Error connecting to message broker server", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("connect", client.name, ""), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error sending connection request", err)
	}
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, resolution, client.logger), nil
}
