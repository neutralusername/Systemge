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
			broker.config.Logger.Warning(Error.New("Failed to accept connection request on broker \""+broker.GetName()+"\"", err).Error())
			continue
		} else {
			broker.config.Logger.Info(Error.New("Accepted connection request from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", nil).Error())
		}
		go broker.handleNodeConnection(netConn)
	}
}

func (broker *Broker) handleNodeConnection(netConn net.Conn) {
	node, err := broker.handleNodeConnectionRequest(netConn)
	if err != nil {
		netConn.Close()
		broker.config.Logger.Warning(Error.New("Failed to handle connection request from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", err).Error())
		return
	} else {
		broker.config.Logger.Info(Error.New("Handled connection request from \""+netConn.RemoteAddr().String()+"\" with name \""+node.name+"\" on broker \""+broker.GetName()+"\"", nil).Error())
	}
	broker.handleNodeConnectionMessages(node)
}

func (broker *Broker) handleNodeConnectionRequest(netConn net.Conn) (*nodeConnection, error) {
	messageBytes, _, err := Utilities.TcpReceive(netConn, broker.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" || message.GetPayload() != "" {
		return nil, Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	nodeConnection := broker.newNodeConnection(message.GetOrigin(), netConn)
	err = broker.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, Error.New("Failed to add node \""+nodeConnection.name+"\"", err)
	}
	err = broker.send(nodeConnection, Message.NewAsync("connected", broker.GetName(), ""))
	if err != nil {
		broker.removeNodeConnection(nodeConnection)
		return nil, Error.New("Failed to send connection response to node \""+nodeConnection.name+"\"", err)
	}
	return nodeConnection, nil
}
