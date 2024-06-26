package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleNodeConnections() {
	for server.IsStarted() {
		netConn, err := server.tlsBrokerListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				server.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go func() {
			node, err := server.handleNodeConnectionRequest(netConn)
			if err != nil {
				netConn.Close()
				server.logger.Log(Error.New("Failed to handle connection request", err).Error())
				return
			}
			server.handleNodeConnectionMessages(node)
		}()
	}
}

func (server *Server) handleNodeConnectionRequest(netConn net.Conn) (*nodeConnection, error) {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return nil, Error.New("Failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" {
		return nil, Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	nodeConnection := newNodeConnection(message.GetOrigin(), netConn)
	err = server.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, err
	}
	err = nodeConnection.send(Message.NewAsync("connected", server.name, ""))
	if err != nil {
		return nil, Error.New("Failed to send connection response to node \""+nodeConnection.name+"\"", err)
	}
	return nodeConnection, nil
}
