package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
)

func (resolver *Resolver) handleResolverConnections() {
	for resolver.isStarted {
		netConn, err := resolver.tlsResolverListener.Accept()
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to accept resolution connection request on resolver \""+resolver.node.GetName()+"\"", err).Error())
			continue
		} else {
			resolver.node.GetLogger().Info(Error.New("Accepted resolution connection request on resolver \""+resolver.node.GetName()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		go resolver.handleResolutionRequest(netConn)
	}
}

func (resolver *Resolver) handleResolutionRequest(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, msgLen, err := Tcp.Receive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.node.GetLogger().Warning(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "failed to receive resolution request").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	if resolver.config.MaxMessageSize > 0 && msgLen > resolver.config.MaxMessageSize {
		resolver.node.GetLogger().Warning(Error.New("Message size exceeds maximum message size from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "message size exceeds maximum message size").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		resolver.node.GetLogger().Warning(Error.New("Invalid connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		resolver.node.GetLogger().Warning(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), Error.New("invalid resolution request", nil).Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	resolver.mutex.Lock()
	endpoint, ok := resolver.registeredTopics[message.GetPayload()]
	resolver.mutex.Unlock()
	if !ok {
		resolver.node.GetLogger().Warning(Error.New("Failed to resolve topic \""+message.GetPayload()+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "unknwon topic").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	err = Tcp.Send(netConn, Message.NewAsync("resolution", resolver.node.GetName(), endpoint.Marshal()).Serialize(), resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.node.GetLogger().Warning(Error.New("Failed to send resolution response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		return
	}
	resolver.node.GetLogger().Info(Error.New("Successfully resolved topic \""+message.GetPayload()+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
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
