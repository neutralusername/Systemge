package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
)

func (node *Node) connectToBroker(tcpEndpoint *TcpEndpoint.TcpEndpoint) (*brokerConnection, error) {
	netConn, err := tcpEndpoint.TlsDial()
	if err != nil {
		return nil, Error.New("Failed connecting to broker", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("connect", node.config.Name, ""), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed sending connection request", err)
	}
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("Invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, tcpEndpoint), nil
}
