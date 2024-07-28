package Resolver

import (
	"net"
	"strings"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (resolver *Resolver) handleConfigConnections() {
	for resolver.isStarted {
		netConn, err := resolver.configTcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept config connection request", err).Error())
			}
			continue
		}
		go func() {
			resolver.configRequestCounter.Add(1)
			if infoLogger := resolver.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Accepted config connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
			if resolver.configTcpServer.GetBlacklist().Contains(ip) {
				netConn.Close()
				if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected config connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
				}
				return
			}
			if resolver.configTcpServer.GetWhitelist().ElementCount() > 0 && !resolver.configTcpServer.GetWhitelist().Contains(ip) {
				netConn.Close()
				if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected config connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
				}
				return
			}
			err := resolver.handleConfigConnection(netConn)
			if err != nil {
				if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle config request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
				bytesSent, err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
				if err != nil {
					if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
					}
				} else {
					resolver.bytesSentCounter.Add(bytesSent)
				}
			} else {
				bytesSent, err := Tcp.Send(netConn, Message.NewAsync("success", resolver.node.GetName(), "").Serialize(), resolver.config.TcpTimeoutMs)
				if err != nil {
					if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
					}
				} else {
					resolver.bytesSentCounter.Add(bytesSent)
				}
			}
			netConn.Close()
		}()
	}
}

func (resolver *Resolver) handleConfigConnection(netConn net.Conn) error {
	messageBytes, bytesReceived, err := Tcp.Receive(netConn, resolver.config.TcpTimeoutMs, resolver.config.IncomingMessageByteLimit)
	if err != nil {
		return Error.New("Failed to receive connection request", err)
	}
	resolver.bytesReceivedCounter.Add(bytesReceived)
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		return Error.New("Invalid connection request", err)
	}
	if message == nil || message.GetOrigin() == "" {
		return Error.New("Invalid connection request", nil)
	}
	err = resolver.handleConfigRequest(message)
	if err != nil {
		return Error.New("Failed to handle config request with topic \""+message.GetTopic()+"\"", err)
	}
	if infoLogger := resolver.node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handled config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
	}
	return nil
}

func (resolver *Resolver) handleConfigRequest(message *Message.Message) error {
	segments := strings.Split(message.GetPayload(), "|")
	switch message.GetTopic() {
	case "addWhitelistResolver":
		for _, segment := range segments {
			resolver.resolverTcpServer.GetWhitelist().Add(segment)
		}
	case "removeWhitelistResolver":
		for _, segment := range segments {
			resolver.resolverTcpServer.GetWhitelist().Remove(segment)
		}
	case "addBlacklistResolver":
		for _, segment := range segments {
			resolver.resolverTcpServer.GetBlacklist().Add(segment)
		}
	case "removeBlacklistResolver":
		for _, segment := range segments {
			resolver.resolverTcpServer.GetBlacklist().Remove(segment)
		}
	case "addWhitelistConfig":
		for _, segment := range segments {
			resolver.configTcpServer.GetWhitelist().Add(segment)
		}
	case "removeWhitelistConfig":
		for _, segment := range segments {
			resolver.configTcpServer.GetWhitelist().Remove(segment)
		}
	case "addBlacklistConfig":
		for _, segment := range segments {
			resolver.configTcpServer.GetBlacklist().Add(segment)
		}
	case "removeBlacklistConfig":
		for _, segment := range segments {
			resolver.configTcpServer.GetBlacklist().Remove(segment)
		}
	case "addTopics":
		if len(segments) < 2 {
			return Error.New("Invalid payload", nil)
		}
		brokerEndpoint := Config.UnmarshalTcpEndpoint(segments[0])
		err := resolver.addTopics(*brokerEndpoint, segments[1:]...)
		if err != nil {
			return Error.New("Failed to add topics", err)
		}
	case "removeTopics":
		if len(segments) < 1 {
			return Error.New("Invalid payload", nil)
		}
		err := resolver.removeTopics(segments...)
		if err != nil {
			return Error.New("Failed to remove topics", err)
		}
	}
	return nil
}
