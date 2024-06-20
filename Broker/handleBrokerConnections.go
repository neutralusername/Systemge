package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleClientConnections() {
	for server.IsStarted() {
		netConn, err := server.tlsBrokerListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				server.logger.Log(Utilities.NewError("Failed to accept connection request", err).Error())
			}
			continue
		}
		go func() {
			client, err := server.handleClientConnectionRequest(netConn)
			if err != nil {
				netConn.Close()
				server.logger.Log(Utilities.NewError("Failed to handle connection request", err).Error())
				return
			}
			server.handleClientConnectionMessages(client)
		}()
	}
}

func (server *Server) handleClientConnectionRequest(netConn net.Conn) (*clientConnection, error) {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return nil, Utilities.NewError("Failed to receive connection request", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" {
		return nil, Utilities.NewError("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	clientConnection := newClientConnection(message.GetOrigin(), netConn)
	err = server.addClientConnection(clientConnection)
	if err != nil {
		return nil, err
	}
	err = clientConnection.send(Message.NewAsync("connected", server.name, ""))
	if err != nil {
		return nil, Utilities.NewError("Failed to send connection response to client \""+clientConnection.name+"\"", err)
	}
	return clientConnection, nil
}
