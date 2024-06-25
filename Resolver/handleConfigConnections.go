package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleConfigConnections() {
	for server.IsStarted() {
		netConn, err := server.tlsConfigListener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				server.logger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		go server.handleConfigConnection(netConn)
	}
}

func (server *Server) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Error.New("failed to receive message", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		server.logger.Log(Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil).Error())
		return
	}
	switch message.GetTopic() {
	case "addKnownBroker":
		resolution := Resolution.Unmarshal(message.GetPayload())
		if resolution == nil {
			err = Error.New("Failed to unmarshal resolution", nil)
			break
		}
		err = server.AddKnownBroker(resolution)
	case "removeKnownBroker":
		err = server.RemoveKnownBroker(message.GetPayload())
	case "addTopics":
		segments := strings.Split(message.GetPayload(), " ")
		if len(segments) < 2 {
			err = Error.New("Invalid payload", nil)
			break
		}
		err = server.AddBrokerTopics(segments[0], segments[1:]...)
	case "removeTopics":
		segments := strings.Split(message.GetPayload(), " ")
		if len(segments) < 1 {
			err = Error.New("Invalid payload", nil)
			break
		}
		server.RemoveBrokerTopics(segments...)
	default:
		err = Error.New("Invalid config request", nil)
	}
	if err != nil {
		server.logger.Log(Error.New("Failed to handle config request \""+message.GetTopic()+"\"", err).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("success", server.name, "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Error.New("Failed to send success message", err).Error())
		return
	}
}
