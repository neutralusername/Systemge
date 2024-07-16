package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (resolver *Resolver) handleConfigConnections() {
	for resolver.IsStarted() {
		netConn, err := resolver.tlsConfigListener.Accept()
		if err != nil {
			resolver.tlsConfigListener.Close()
			if resolver.IsStarted() {
				resolver.logger.Info(Error.New("Failed to accept connection request", err).Error())
			}
			return
		}
		go resolver.handleConfigConnection(netConn)
	}
}

func (resolver *Resolver) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, msgLen, err := Utilities.TcpReceive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.logger.Info(Error.New("failed to receive message", err).Error())
		return
	}
	if resolver.config.MaxMessageSize > 0 && msgLen > resolver.config.MaxMessageSize {
		resolver.logger.Warning(Error.New("Message exceeds maximum message size from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", nil).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", resolver.GetName(), "message size exceeds maximum message size").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		}
		return
	}
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		resolver.logger.Warning(Error.New("Invalid connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", resolver.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		}
		return
	}
	if message == nil || message.GetOrigin() == "" {
		resolver.logger.Info(Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil).Error())
		return
	}
	switch message.GetTopic() {
	case "addTopics":
		segments := strings.Split(message.GetPayload(), "|")
		if len(segments) < 2 {
			err = Error.New("Invalid payload \""+message.GetPayload()+"\"", nil)
			break
		}
		brokerEndpoint := TcpEndpoint.Unmarshal(segments[0])
		for _, topic := range segments[1:] {
			err = resolver.AddTopic(*brokerEndpoint, topic)
			if err != nil {
				resolver.logger.Info(Error.New("Failed to add topic \""+topic+"\"", err).Error())
			}
		}
	case "removeTopics":
		segments := strings.Split(message.GetPayload(), "|")
		if len(segments) < 1 {
			err = Error.New("Invalid payload", nil)
			break
		}
		for _, topic := range segments {
			err = resolver.RemoveTopic(topic)
			if err != nil {
				resolver.logger.Info(Error.New("Failed to remove topic \""+topic+"\"", err).Error())
			}
		}
	default:
		err = Error.New("Invalid config request", nil)
	}
	if err != nil {
		resolver.logger.Info(Error.New("Failed to handle config request \""+message.GetTopic()+"\"", err).Error())
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("success", resolver.GetName(), "").Serialize(), resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.logger.Info(Error.New("Failed to send success message", err).Error())
		return
	}
}
