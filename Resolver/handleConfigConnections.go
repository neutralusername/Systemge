package Resolver

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (server *Server) handleConfigConnections(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		server.logger.Log(Utilities.NewError("failed to receive message", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "config" || message.GetOrigin() == "" {
		server.logger.Log(Utilities.NewError("Invalid connection request \""+string(messageBytes)+"\"", nil).Error())
		return
	}
	switch message.GetPayload() {
	case "registerBroker":
		resolution := UnmarshalResolution(message.GetPayload())
		if resolution == nil {
			err = Utilities.NewError("Failed to unmarshal resolution", nil)
			break
		}
		err = server.RegisterBroker(resolution)
	case "unregisterBroker":
		err = server.UnregisterBroker(message.GetPayload())
	case "registerTopics":
		segments := strings.Split(message.GetPayload(), " ")
		if len(segments) < 2 {
			err = Utilities.NewError("Invalid payload", nil)
			break
		}
		err = server.RegisterTopics(segments[0], segments[1:]...)
	case "unregisterTopics":
		segments := strings.Split(message.GetPayload(), " ")
		if len(segments) < 1 {
			err = Utilities.NewError("Invalid payload", nil)
			break
		}
		server.UnregisterTopic(segments...)
	default:
		err = Utilities.NewError("Invalid config request", nil)
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
