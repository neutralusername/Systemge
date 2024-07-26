package Resolver

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
)

func (resolver *Resolver) handleResolverConnections() {
	for resolver.isStarted {
		netConn, err := resolver.resolverTcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept resolution connection request", err).Error())
			}
			continue
		}
		resolver.resolutionRequestCounter.Add(1)
		if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Accepted resolution connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
		if resolver.resolverTcpServer.GetBlacklist().Contains(ip) {
			netConn.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected resolution connection request from \""+netConn.RemoteAddr().String()+"\" due to blacklist", nil).Error())
			}
			continue
		}
		if resolver.resolverTcpServer.GetWhitelist().ElementCount() > 0 && !resolver.resolverTcpServer.GetWhitelist().Contains(ip) {
			netConn.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected resolution connection request from \""+netConn.RemoteAddr().String()+"\" due to whitelist", nil).Error())
			}
			continue
		}
		go func() {
			err := resolver.handleResolutionRequest(netConn)
			if err != nil {
				if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle resolution request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
				bytesSent, err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
				if err != nil {
					if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send error response to resolver connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
					}
				} else {
					resolver.bytesSentCounter.Add(bytesSent)
				}
			}
		}()
	}
}

func (resolver *Resolver) handleResolutionRequest(netConn net.Conn) error {
	defer netConn.Close()
	messageBytes, bytesReceived, err := Tcp.Receive(netConn, resolver.config.TcpTimeoutMs, resolver.config.IncomingMessageByteLimit)
	if err != nil {
		return Error.New("failed to receive resolution request", err)
	}
	resolver.bytesReceivedCounter.Add(bytesReceived)
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		return Error.New("invalid resolution request", err)
	}
	if message == nil || message.GetTopic() != "resolve" || message.GetOrigin() == "" {
		return Error.New("invalid resolution request", nil)
	}
	resolver.mutex.Lock()
	endpoint, ok := resolver.registeredTopics[message.GetPayload()]
	resolver.mutex.Unlock()
	if !ok {
		return Error.New("unknown topic", nil)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync("resolution", resolver.node.GetName(), endpoint.Marshal()).Serialize(), resolver.config.TcpTimeoutMs)
	if err != nil {
		return Error.New("failed to send resolution response", err)
	}
	resolver.bytesSentCounter.Add(bytesSent)
	if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Resolved topic \""+message.GetPayload()+"\" to \""+endpoint.Address+"\" from resolver connection \""+netConn.RemoteAddr().String()+"\"", nil).Error())
	}
	return nil
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
