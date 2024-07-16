package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
)

func (resolver *Resolver) handleResolverConnections() {
	for resolver.IsStarted() {
		netConn, err := resolver.tlsResolverListener.Accept()
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to accept connection request on resolver \""+resolver.GetName()+"\"", err).Error())
			continue
		} else {
			resolver.logger.Info(Error.New("Accepted connection request on resolver \""+resolver.GetName()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		go resolver.handleResolutionRequest(netConn)
	}
}

func (resolver *Resolver) handleResolutionRequest(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, msgLen, err := Utilities.TcpReceive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.logger.Warning(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", resolver.GetName(), "failed to receive resolution request").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		}
		return
	}
	if resolver.config.MaxMessageSize > 0 && msgLen > resolver.config.MaxMessageSize {
		resolver.logger.Warning(Error.New("Message size exceeds maximum message size from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", nil).Error())
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
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		resolver.logger.Warning(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", nil).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", resolver.GetName(), Error.New("invalid resolution request", nil).Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		}
		return
	}
	resolver.mutex.Lock()
	endpoint, ok := resolver.registeredTopics[message.GetPayload()]
	resolver.mutex.Unlock()
	if !ok {
		resolver.logger.Warning(Error.New("Failed to resolve topic \""+message.GetPayload()+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", nil).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", resolver.GetName(), "unknwon topic").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
			return
		}
		return
	}
	err = Utilities.TcpSend(netConn, Message.NewAsync("resolution", resolver.GetName(), endpoint.Marshal()).Serialize(), resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.logger.Warning(Error.New("Failed to send resolution response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		return
	}
	resolver.logger.Info(Error.New("Successfully resolved topic \""+message.GetPayload()+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", nil).Error())
}

func (resolver *Resolver) validateMessage(message *Message.Message) error {
	if message.GetSyncRequestToken() != "" {
		return Error.New("synchronous request token is not empty", nil)
	}
	if message.GetSyncResponseToken() != "" {
		return Error.New("synchronous response token is not empty", nil)
	}
	if resolver.config.MaxOriginSize > 0 && len(message.GetOrigin()) > resolver.config.MaxOriginSize {
		return Error.New("origin size exceeds maximum origin size", nil)
	}
	if resolver.config.MaxPayloadSize > 0 && len(message.GetPayload()) > (resolver.config.MaxPayloadSize) {
		return Error.New("payload size exceeds maximum payload size", nil)

	}
	if resolver.config.MaxTopicSize > 0 && len(message.GetTopic()) > resolver.config.MaxTopicSize {
		return Error.New("topic size exceeds maximum topic size", nil)
	}
	return nil
}
