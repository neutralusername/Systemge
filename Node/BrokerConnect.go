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
		return nil, Error.New("failed connecting to broker", err)
	}
	responseMessage, err := Utilities.TcpExchange(netConn, Message.NewAsync("connect", node.config.Name, ""), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("failed sending connection request", err)
	}
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, tcpEndpoint, node.logger), nil
}
