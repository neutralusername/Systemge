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
		netConn, err := broker.configTcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			continue
		}
		ip, _, err := net.SplitHostPort(netConn.RemoteAddr().String())
		if err != nil {
			netConn.Close()
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to get remote address", err).Error())
			}
			continue
		}
		if broker.configTcpServer.GetBlacklist().Contains(ip) {
			netConn.Close()
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			continue
		}
		if broker.configTcpServer.GetWhitelist().ElementCount() > 0 && !broker.configTcpServer.GetWhitelist().Contains(ip) {
			netConn.Close()
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			continue
		}
		if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Accepted connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		go broker.handleConfigConnection(netConn)
	}
}

func (broker *Broker) handleConfigConnection(netConn net.Conn) {
	defer netConn.Close()
	messageBytes, _, err := Tcp.Receive(netConn, broker.config.TcpTimeoutMs, broker.config.IncomingMessageByteLimit)
	if err != nil {
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to receive connection request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		return
	}
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetOrigin() == "" {
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Invalid connection request \""+string(messageBytes)+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		return
	}
	err = broker.validateMessage(message)
	if err != nil {
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Invalid connection request message from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		return
	}
	err = broker.handleConfigRequest(message)
	if err != nil {
		if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed to handle connection request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("error", broker.node.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), broker.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
	} else {
		if infoLogger := broker.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Handled config request with topic \""+message.GetTopic()+"\" from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
		}
		err := Tcp.Send(netConn, Message.NewAsync("success", broker.node.GetName(), "").Serialize(), broker.config.TcpTimeoutMs)
		if err != nil {
			if warningLogger := broker.node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
			}
		}
	}
}

func (broker *Broker) handleConfigRequest(message *Message.Message) error {
	payloadSegments := strings.Split(message.GetPayload(), "|")
	if len(payloadSegments) == 0 {
		return Error.New("No topics provided", nil)
	}
	switch message.GetTopic() {
	case "addWhitelistBroker":
		for _, payloadSegment := range payloadSegments {
			broker.brokerTcpServer.GetWhitelist().Add(payloadSegment)
		}
	case "removeWhitelistBroker":
		for _, payloadSegment := range payloadSegments {
			broker.brokerTcpServer.GetWhitelist().Remove(payloadSegment)
		}
	case "addBlacklistBroker":
		for _, payloadSegment := range payloadSegments {
			broker.brokerTcpServer.GetBlacklist().Add(payloadSegment)
		}
	case "removeBlacklistBroker":
		for _, payloadSegment := range payloadSegments {
			broker.brokerTcpServer.GetBlacklist().Remove(payloadSegment)
		}
	case "addWhitelistConfig":
		for _, payloadSegment := range payloadSegments {
			broker.configTcpServer.GetWhitelist().Add(payloadSegment)
		}
	case "removeWhitelistConfig":
		for _, payloadSegment := range payloadSegments {
			broker.configTcpServer.GetWhitelist().Remove(payloadSegment)
		}
	case "addBlacklistConfig":
		for _, payloadSegment := range payloadSegments {
			broker.configTcpServer.GetBlacklist().Add(payloadSegment)
		}
	case "removeBlacklistConfig":
		for _, payloadSegment := range payloadSegments {
			broker.configTcpServer.GetBlacklist().Remove(payloadSegment)
		}
	case "addSyncTopics":
		broker.addSyncTopics(payloadSegments...)
		err := broker.addResolverTopicsRemotely(payloadSegments...)
		if err != nil {
			return Error.New("Failed to add topics remotely", err)
		}
	case "removeSyncTopics":
		broker.removeSyncTopics(payloadSegments...)
		err := broker.removeResolverTopicsRemotely(payloadSegments...)
		if err != nil {
			return Error.New("Failed to remove topics remotely", err)
		}
	case "addAsyncTopics":
		broker.addAsyncTopics(payloadSegments...)
		err := broker.addResolverTopicsRemotely(payloadSegments...)
		if err != nil {
			return Error.New("Failed to add topics remotely", err)
		}
	case "removeAsyncTopics":
		broker.removeAsyncTopics(payloadSegments...)
		err := broker.removeResolverTopicsRemotely(payloadSegments...)
		if err != nil {
			return Error.New("Failed to remove topics remotely", err)
		}
	default:
		return Error.New("Unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
