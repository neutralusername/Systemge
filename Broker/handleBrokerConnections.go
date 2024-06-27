package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (broker *Broker) handleNodeConnections() {
	for broker.IsStarted() {
		netConn, err := broker.tlsBrokerListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				broker.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go func() {
			node, err := broker.handleNodeConnectionRequest(netConn)
			if err != nil {
				netConn.Close()
				broker.logger.Log(Error.New("Failed to handle connection request", err).Error())
				return
			}
			broker.handleNodeConnectionMessages(node)
		}()
	}
}

func (broker *Broker) handleNodeConnectionRequest(netConn net.Conn) (*nodeConnection, error) {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return nil, Error.New("Failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" {
		return nil, Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	nodeConnection := newNodeConnection(message.GetOrigin(), netConn)
	err = broker.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, err
	}
	err = nodeConnection.send(Message.NewAsync("connected", broker.name, ""))
	if err != nil {
		return nil, Error.New("Failed to send connection response to node \""+nodeConnection.name+"\"", err)
	}
	return nodeConnection, nil
}
