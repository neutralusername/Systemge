package Resolver

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
	"strings"
)

func (resolver *Resolver) handleConfigConnections() {
	for resolver.isStarted {
		netConn, err := resolver.tlsConfigListener.Accept()
		if err != nil {
			resolver.tlsConfigListener.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept config connection request", err).Error())
			}
			continue
		}
		if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Accepted config connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		if err := resolver.validateAddressConfig(netConn.RemoteAddr().String()); err != nil {
			netConn.Close()
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(err.Error())
			}
			continue
		}
		go resolver.handleConfigConnection(netConn)
	}
}

func (resolver *Resolver) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, msgLen, err := Tcp.Receive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to receive message from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		return
	}
	if resolver.config.MaxMessageSize > 0 && msgLen > resolver.config.MaxMessageSize {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Message exceeds maximum message size from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "message size exceeds maximum message size").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	if message == nil || message.GetOrigin() == "" {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Invalid connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), Error.New("invalid resolution request", nil).Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
		return
	}
	err = resolver.handleConfigRequest(message)
	if err != nil {
		if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to handle config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
	} else {
		if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Handled config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("success", resolver.node.GetName(), "Handled request").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := resolver.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
	}
}

func (resolver *Resolver) handleConfigRequest(message *Message.Message) error {
	segments := strings.Split(message.GetPayload(), "|")
	switch message.GetTopic() {
	case "addWhitelistResolver":
		resolver.addToresolverWhitelist(segments...)
	case "removeWhitelistResolver":
		resolver.removeFromresolverWhitelist(segments...)
	case "addBlacklistResolver":
		resolver.addToresolverBlacklist(segments...)
	case "removeBlacklistResolver":
		resolver.removeFromresolverBlacklist(segments...)
	case "addWhitelistConfig":
		resolver.addToConfigWhitelist(segments...)
	case "removeWhitelistConfig":
		resolver.removeFromConfigWhitelist(segments...)
	case "addBlacklistConfig":
		resolver.addToConfigBlacklist(segments...)
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
