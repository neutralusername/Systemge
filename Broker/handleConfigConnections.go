package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"strings"
)

func (broker *Broker) handleConfigConnections() {
	for broker.IsStarted() {
		netConn, err := broker.tlsConfigListener.Accept()
		if err != nil {
			broker.config.Logger.Warning(Error.New("Failed to accept connection request on broker \""+broker.GetName()+"\"", err).Error())
			continue
		} else {
			broker.config.Logger.Info(Error.New("Accepted connection request on broker \""+broker.GetName()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		go broker.handleConfigConnection(netConn)
	}
}

func (broker *Broker) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		broker.config.Logger.Warning(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", err).Error())
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		broker.config.Logger.Warning(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", nil).Error())
		return
	}
	err = broker.handleConfigRequest(message)
	if err != nil {
		broker.config.Logger.Warning(Error.New("Failed to handle config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", err).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("error", broker.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), DEFAULT_TCP_TIMEOUT)
		if err != nil {
			broker.config.Logger.Warning(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", err).Error())
		}
	} else {
		broker.config.Logger.Info(Error.New("Successfully handled config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", nil).Error())
		err := Utilities.TcpSend(netConn, Message.NewAsync("success", broker.GetName(), "").Serialize(), DEFAULT_TCP_TIMEOUT)
		if err != nil {
			broker.config.Logger.Warning(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\" on broker \""+broker.GetName()+"\"", err).Error())
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
