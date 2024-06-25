package Client

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (client *Client) connectToBroker(resolution *Resolution.Resolution) (*brokerConnection, error) {
	netConn, err := Utilities.TlsDial(resolution.GetAddress(), resolution.GetServerNameIndication(), resolution.GetTlsCertificate())
	if err != nil {
		return nil, Error.New("Error connecting to message broker server", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("connect", client.config.Name, ""), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error sending connection request", err)
	}
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("Invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, resolution, client.logger), nil
}
