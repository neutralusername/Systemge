package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
)

func (server *Server) handleConfigConnections() {
	for server.IsStarted() {
		for server.IsStarted() {
			netConn, err := server.tlsConfigListener.Accept()
			if err != nil {
				if server.IsStarted() {
					server.logger.Log(Utilities.NewError("Failed to accept connection", err).Error())
				}
				return
			}
			go server.handleConfigConnection(netConn)
		}
	}
}

func (server *Server) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Utilities.NewError("failed to receive message", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		server.logger.Log(Utilities.NewError("Invalid connection request \""+string(messageBytes)+"\"", nil).Error())
		return
	}
	switch message.GetTopic() {

	}
	if err != nil {
		server.logger.Log(Utilities.NewError("Failed to handle config request", err).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("success", server.name, "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Utilities.NewError("Failed to send success message", err).Error())
		return
	}

}
