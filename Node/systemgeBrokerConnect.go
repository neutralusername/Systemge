package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
)

func (systemge *systemgeComponent) connectToBroker(nodeName string, tcpEndpoint *Config.TcpEndpoint) (*brokerConnection, error) {
	netConn, err := Tcp.NewEndpoint(tcpEndpoint)
	if err != nil {
		return nil, Error.New("Failed connecting to broker", err)
	}
	messageBytes := Message.NewAsync("connect", nodeName, "").Serialize()
	responseMessage, bytesReceived, err := Tcp.Exchange(netConn, messageBytes, systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, 0)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed sending connection request", err)
	}
	systemge.bytesSentCounter.Add(uint64(len(messageBytes)))
	systemge.bytesReceivedCounter.Add(uint64(bytesReceived))
	if responseMessage.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("Invalid response topic \""+responseMessage.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, tcpEndpoint), nil
}
