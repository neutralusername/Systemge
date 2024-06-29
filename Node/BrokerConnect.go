package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (node *Node) connectToBroker(resolution *Resolution.Resolution) (*brokerConnection, error) {
	netConn, err := Utilities.TlsDial(resolution.GetAddress(), resolution.GetServerNameIndication(), resolution.GetTlsCertificate())
	if err != nil {
		return nil, Error.New("Error connecting to broker", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("connect", node.config.Name, ""), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error sending connection request", err)
	}
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("Invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, resolution, node.logger), nil
}
