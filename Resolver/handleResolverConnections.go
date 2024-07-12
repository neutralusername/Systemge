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
	messageBytes, err := Utilities.TcpReceive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.logger.Warning(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", resolver.GetName(), "failed to receive resolution request").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.logger.Warning(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.GetName()+"\"", err).Error())
		}
		return
	}
	message := Message.Deserialize(messageBytes)
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
