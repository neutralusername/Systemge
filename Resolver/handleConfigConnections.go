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
			resolver.node.GetLogger().Warning(Error.New("Failed to accept config connection request on resolver \""+resolver.node.GetName()+"\"", err).Error())
			continue
		} else {
			resolver.node.GetLogger().Info(Error.New("Accepted config connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		}
		go resolver.handleConfigConnection(netConn)
	}
}

func (resolver *Resolver) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, msgLen, err := Tcp.Receive(netConn, resolver.config.TcpTimeoutMs)
	if err != nil {
		resolver.node.GetLogger().Warning(Error.New("failed to receive message from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		return
	}
	if resolver.config.MaxMessageSize > 0 && msgLen > resolver.config.MaxMessageSize {
		resolver.node.GetLogger().Warning(Error.New("Message exceeds maximum message size from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), "message size exceeds maximum message size").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	message := Message.Deserialize(messageBytes)
	err = resolver.validateMessage(message)
	if err != nil {
		resolver.node.GetLogger().Warning(Error.New("Invalid connection request from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	if message == nil || message.GetOrigin() == "" {
		resolver.node.GetLogger().Warning(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), Error.New("invalid resolution request", nil).Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
		return
	}
	err = resolver.handleConfigRequest(message)
	if err != nil {
		resolver.node.GetLogger().Warning(Error.New("Failed to handle config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", resolver.node.GetName(), err.Error()).Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
	} else {
		resolver.node.GetLogger().Info(Error.New("Handled config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("success", resolver.node.GetName(), "Handled request").Serialize(), resolver.config.TcpTimeoutMs)
		if err != nil {
			resolver.node.GetLogger().Warning(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\" on resolver \""+resolver.node.GetName()+"\"", err).Error())
		}
	}
}

func (resolver *Resolver) handleConfigRequest(message *Message.Message) error {
	switch message.GetTopic() {
	case "addTopics":
		segments := strings.Split(message.GetPayload(), "|")
		if len(segments) < 2 {
			return Error.New("Invalid payload", nil)
		}
		brokerEndpoint := Config.UnmarshalTcpEndpoint(segments[0])
		err := resolver.addTopics(*brokerEndpoint, segments[1:]...)
		if err != nil {
			return Error.New("Failed to add topics", err)
		}
	case "removeTopics":
		segments := strings.Split(message.GetPayload(), "|")
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
