package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleResolverConnections() {
	for server.IsStarted() {
		netConn, err := server.tlsResolverListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				server.logger.Log(Error.New("Failed to accept connection request", err).Error())
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
		server.logger.Log(Error.New("", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		server.logger.Log(Error.New("Invalid connection request", nil).Error())
		return
	}
	server.mutex.Lock()
	broker, ok := server.registeredTopics[message.GetPayload()]
	server.mutex.Unlock()
	if !ok {
		server.logger.Log(Error.New("Topic not found: \""+message.GetPayload()+"\" by "+message.GetOrigin(), nil).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("resolution", server.name, broker.resolution.Marshal()).Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Error.New("Failed to send resolution to \""+message.GetOrigin()+"\"", err).Error())
		return
	}
}
