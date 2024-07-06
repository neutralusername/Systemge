package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
)

func (broker *Broker) handleNodeConnections() {
	for broker.IsStarted() {
		netConn, err := broker.tlsBrokerListener.Accept()
		if err != nil {
			broker.tlsBrokerListener.Close()
			if broker.IsStarted() {
				broker.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			return
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
	nodeConnection := broker.newNodeConnection(message.GetOrigin(), netConn)
	err = broker.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, Error.New("Failed to add node connection", err)
	}
	broker.startWatchdog(nodeConnection)
	err = nodeConnection.send(Message.NewAsync("connected", broker.GetName(), ""))
	if err != nil {
		return nil, Error.New("Failed to send connection response to node \""+nodeConnection.name+"\"", err)
	}
	return nodeConnection, nil
}
