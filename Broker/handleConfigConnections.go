package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tcp"
	"net"
	"strings"
)

func (broker *Broker) handleConfigConnections() {
	for broker.isStarted {
		netConn, err := broker.tlsConfigListener.Accept()
		if err != nil {
			broker.node.GetLogger().Warning(Error.New("Failed to accept connection request on broker \""+broker.node.GetName()+"\"", err).Error())
			continue
		}
		broker.node.GetLogger().Info(Error.New("Accepted connection request on broker \""+broker.node.GetName()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		go broker.handleConfigConnection(netConn)
	}
}

func (broker *Broker) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, _, err := Tcp.Receive(netConn, broker.config.TcpTimeoutMs)
	if err != nil {
		broker.node.GetLogger().Warning(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		broker.node.GetLogger().Warning(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", nil).Error())
		return
	}
	err = broker.validateMessage(message)
	if err != nil {
		broker.node.GetLogger().Warning(Error.New("Invalid connection request message from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", err).Error())
		return
	}
	err = broker.handleConfigRequest(message)
	if err != nil {
		broker.node.GetLogger().Warning(Error.New("Failed to handle config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", err).Error())
		err := Tcp.Send(netConn, Message.NewAsync("error", broker.node.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), broker.config.TcpTimeoutMs)
		if err != nil {
			broker.node.GetLogger().Warning(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", err).Error())
		}
	} else {
		broker.node.GetLogger().Info(Error.New("Successfully handled config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", nil).Error())
		err := Tcp.Send(netConn, Message.NewAsync("success", broker.node.GetName(), "").Serialize(), broker.config.TcpTimeoutMs)
		if err != nil {
			broker.node.GetLogger().Warning(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.node.GetName()+"\"", err).Error())
		}
	}
}

func (broker *Broker) handleConfigRequest(message *Message.Message) error {
	topics := strings.Split(message.GetPayload(), "|")
	if len(topics) == 0 {
		return Error.New("No topics provided", nil)
	}
	switch message.GetTopic() {
	case "addSyncTopics":
		broker.addSyncTopics(topics...)
		err := broker.addResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("Failed to add topics remotely", err)
		}
	case "removeSyncTopics":
		broker.removeSyncTopics(topics...)
		err := broker.removeResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("Failed to remove topics remotely", err)
		}
	case "addAsyncTopics":
		broker.addAsyncTopics(topics...)
		err := broker.addResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("Failed to add topics remotely", err)
		}
	case "removeAsyncTopics":
		broker.removeAsyncTopics(topics...)
		err := broker.removeResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("Failed to remove topics remotely", err)
		}
	default:
		return Error.New("Unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
