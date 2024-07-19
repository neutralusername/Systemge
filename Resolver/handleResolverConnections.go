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
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept resolution connection request", err).Error())
			}
			continue
		}
		if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Accepted resolution connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		ip, _, err := net.SplitHostPort(netConn.RemoteAddr().String())
		if err != nil {
			netConn.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to get remote address", err).Error())
			}
			continue
		}
		if resolver.resolverBlacklist.Contains(ip) {
			netConn.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected resolution connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			continue
		}
		if resolver.resolverWhitelist.ElementCount() > 0 && !resolver.resolverWhitelist.Contains(ip) {
			netConn.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected resolution connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			continue
		}
		go resolver.handleResolutionRequest(netConn)
	}
}

func (resolver *Resolver) handleResolutionRequest(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, msgLen, err := Tcp.Receive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "failed to receive resolution request").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	if resolver.config.MaxMessageSize > 0 && msgLen > resolver.config.MaxMessageSize {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Message size exceeds maximum message size from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "message size exceeds maximum message size").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Invalid connection request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), Error.New("invalid resolution request", nil).Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	resolver.mutex.Lock()
	endpoint, ok := resolver.registeredTopics[message.GetPayload()]
	resolver.mutex.Unlock()
	if !ok {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to resolve topic \""+message.GetPayload()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "unknwon topic").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	err = Tcp.Send(netConn, Message.NewAsync("resolution", resolver.node.GetName(), endpoint.Marshal()).Serialize(), resolver.config.TcpTimeoutMs)
	if err != nil {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to send resolution response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		return
	}
	if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Resolved topic \""+message.GetPayload()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
	}
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
