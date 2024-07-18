package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
)

func (broker *Broker) handleNodeConnections() {
	for broker.isStarted {
		netConn, err := broker.tlsBrokerListener.Accept()
		if err != nil {
			broker.node.GetLogger().Warning(Error.New("Failed to accept connection request on broker \""+broker.node.GetName()+"\"", err).Error())
			continue
		}
		broker.node.GetLogger().Info(Error.New("Accepted connection request from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", nil).Error())
		go broker.handleNodeConnection(netConn)
	}
}

func (broker *Broker) handleNodeConnection(netConn net.Conn) {
	nodeConnection, err := broker.handleNodeConnectionRequest(netConn)
	if err != nil {
		netConn.Close()
		broker.node.GetLogger().Warning(Error.New("Failed to handle connection request from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", err).Error())
		return
	}
	broker.node.GetLogger().Info(Error.New("Handled connection request from \""+netConn.RemoteAddr().String()+"\" with name \""+nodeConnection.name+"\" on broker \""+broker.node.GetName()+"\"", nil).Error())
	broker.handleNodeConnectionMessages(nodeConnection)
	broker.removeNodeConnection(true, nodeConnection)
}

func (broker *Broker) handleNodeConnectionRequest(netConn net.Conn) (*nodeConnection, error) {
	messageBytes, _, err := Tcp.Receive(netConn, broker.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" || message.GetPayload() != "" {
		return nil, Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	err = broker.validateMessage(message)
	if err != nil {
		return nil, Error.New("Invalid connection request message", err)
	}
	nodeConnection := broker.newNodeConnection(message.GetOrigin(), netConn)
	err = broker.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, Error.New("Failed to add node \""+nodeConnection.name+"\"", err)
	}
	err = broker.send(nodeConnection, Message.NewAsync("connected", broker.node.GetName(), ""))
	if err != nil {
		broker.removeNodeConnection(true, nodeConnection)
		return nil, Error.New("Failed to send connection response to node \""+nodeConnection.name+"\"", err)
	}
	return nodeConnection, nil
}
