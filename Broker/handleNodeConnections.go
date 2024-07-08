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
			if broker.IsStarted() {
				broker.logger.Log(Error.New("failed to accept connection request on broker \""+broker.GetName()+"\"", err).Error())
			}
			continue
		}
		go func() {
			node, err := broker.handleNodeConnectionRequest(netConn)
			if err != nil {
				netConn.Close()
				broker.logger.Log(Error.New("failed to handle connection request on broker \""+broker.GetName()+"\"", err).Error())
				return
			}
			broker.handleNodeConnectionMessages(node)
		}()
	}
}

func (broker *Broker) handleNodeConnectionRequest(netConn net.Conn) (*nodeConnection, error) {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return nil, Error.New("failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" {
		return nil, Error.New("invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	nodeConnection := broker.newNodeConnection(message.GetOrigin(), netConn)
	err = broker.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, Error.New("failed to add node \""+nodeConnection.name+"\"", err)
	}
	err = nodeConnection.send(Message.NewAsync("connected", broker.GetName(), ""))
	if err != nil {
		errRemove := broker.removeNodeConnection(nodeConnection)
		if errRemove != nil {
			broker.logger.Log(Error.New("failed to remove node \""+nodeConnection.name+"\" after failed connection response on broker \""+broker.GetName()+"\"", errRemove).Error())
		}
		return nil, Error.New("failed to send connection response to node \""+nodeConnection.name+"\"", err)
	}
	return nodeConnection, nil
}
