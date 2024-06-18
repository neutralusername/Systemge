package Resolver

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleResolverConnections() {
	for server.IsStarted() {
		netConn, err := server.tcpListenerResolver.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				server.logger.Log(Utilities.NewError("Failed to accept connection request", err).Error())
			}
			continue
		}
		go server.handleResolverConnection(netConn)
	}
}

func (server *Server) handleResolverConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Utilities.NewError("", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		server.logger.Log(Utilities.NewError("Invalid connection request", nil).Error())
		return
	}
	server.mutex.Lock()
	broker, ok := server.registeredTopics[message.GetPayload()]
	server.mutex.Unlock()
	if !ok {
		server.logger.Log(Utilities.NewError("Topic not found: \""+message.GetPayload()+"\" by "+message.GetOrigin(), nil).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("resolution", server.name, broker.Marshal()).Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Utilities.NewError("Failed to send resolution to \""+message.GetOrigin()+"\"", err).Error())
		return
	}
}
