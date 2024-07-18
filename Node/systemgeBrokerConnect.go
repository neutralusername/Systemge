package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
)

func (node *Node) connectToBroker(tcpEndpoint Config.TcpEndpoint) (*brokerConnection, error) {
	netConn, err := Tcp.NewClient(tcpEndpoint)
	if err != nil {
		return nil, Error.New("Failed connecting to broker", err)
	}
	responseMessage, err := Tcp.Exchange(netConn, Message.NewAsync("connect", node.GetName(), ""), node.GetSystemgeComponent().GetSystemgeComponentConfig().TcpTimeoutMs)
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
