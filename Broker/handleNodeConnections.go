package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
)

func (broker *Broker) handleNodeConnections() {
	for broker.isStarted {
		netConn, err := broker.brokerTcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		ip, _, err := net.SplitHostPort(netConn.RemoteAddr().String())
		if err != nil {
			netConn.Close()
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to get remote address", err).Error())
			}
			continue
		}
		if broker.brokerTcpServer.GetBlacklist().Contains(ip) {
			netConn.Close()
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			continue
		}
		if broker.brokerTcpServer.GetWhitelist().ElementCount() > 0 && !broker.brokerTcpServer.GetWhitelist().Contains(ip) {
			netConn.Close()
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			continue
		}
		if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Accepted connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		go broker.handleNodeConnection(netConn)
	}
}

func (broker *Broker) handleNodeConnection(netConn net.Conn) {
	nodeConnection, err := broker.handleNodeConnectionRequest(netConn)
	if err != nil {
		netConn.Close()
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to handle connection request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		return
	}
	if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handled connection request from \""+netConn.RemoteAddr().String()+"\" with name \""+nodeConnection.name+"\"", nil).Error())
	}
	broker.handleNodeConnectionMessages(nodeConnection)
	broker.removeNodeConnection(true, nodeConnection)
}

func (broker *Broker) handleNodeConnectionRequest(netConn net.Conn) (*nodeConnection, error) {
	messageBytes, _, err := Tcp.Receive(netConn, broker.config.TcpTimeoutMs, broker.config.IncomingMessageByteLimit)
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
	bytesSent, err := Tcp.Send(nodeConnection.netConn, Message.NewAsync("connected", broker.node.GetName(), "").Serialize(), broker.config.TcpTimeoutMs)
	if err != nil {
		broker.removeNodeConnection(true, nodeConnection)
		return nil, Error.New("Failed to send message", err)
	}
	broker.bytesSentCounter.Add(bytesSent)
	return nodeConnection, nil
}
