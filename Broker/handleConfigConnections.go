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
			if broker.IsStarted() {
				broker.logger.Log(Error.New("failed to accept connection request on broker \""+broker.GetName()+"\"", err).Error())
			}
			continue
		}
		go func() {
			defer netConn.Close()
			err := broker.handleConfigConnection(netConn)
			if err != nil {
				errSend := Utilities.TcpSend(netConn, Message.NewAsync("error", broker.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), DEFAULT_TCP_TIMEOUT)
				if errSend != nil {
					broker.logger.Log(Error.New("failed to send error response to config connection on broker \""+broker.GetName()+"\"", errSend).Error())
				}
				broker.logger.Log(Error.New("failed to handle config request on broker \""+broker.GetName()+"\"", err).Error())
			} else {
				errSend := Utilities.TcpSend(netConn, Message.NewAsync("success", broker.GetName(), "").Serialize(), DEFAULT_TCP_TIMEOUT)
				if errSend != nil {
					broker.logger.Log(Error.New("failed to send success response to config connection on broker \""+broker.GetName()+"\"", errSend).Error())
				}
				broker.logger.Log(Error.New("successfully handled config request on broker \""+broker.GetName()+"\"", nil).Error())
			}
		}()
	}
}

func (broker *Broker) handleConfigConnection(netConn net.Conn) error {
	messageBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("failed to receive message", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		return Error.New("invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	topics := strings.Split(message.GetPayload(), "|")
	if len(topics) == 0 {
		return Error.New("no topics provided", nil)
	}
	switch message.GetTopic() {
	case "addSyncTopics":
		broker.addSyncTopics(topics...)
		err := broker.addResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("failed to add topics remotely", err)
		}
	case "removeSyncTopics":
		broker.removeSyncTopics(topics...)
		err := broker.removeResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("failed to remove topics remotely", err)
		}
	case "addAsyncTopics":
		broker.addAsyncTopics(topics...)
		err := broker.addResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("failed to add topics remotely", err)
		}
	case "removeAsyncTopics":
		broker.removeAsyncTopics(topics...)
		err := broker.removeResolverTopicsRemotely(topics...)
		if err != nil {
			return Error.New("failed to remove topics remotely", err)
		}
	default:
		return Error.New("unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
