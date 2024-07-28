package Node

import (
	"net"
	"strings"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (node *Node) handleBrokerConfigConnections() {
	broker_ := node.broker
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handling broker config connections", nil).Error())
	}
	defer func() {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Stopped handling broker config connections", nil).Error())
		}
	}()
	for {
		broker := node.broker
		if broker != broker_ {
			return
		}
		netConn, err := broker.configTcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			return
		}
		go func() {
			broker.configRequestCounter.Add(1)
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Accepted config request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
			if broker.configTcpServer.GetBlacklist().Contains(ip) {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\" due to blacklist", nil).Error())
				}
				return
			}
			if broker.configTcpServer.GetWhitelist().ElementCount() > 0 && !broker.configTcpServer.GetWhitelist().Contains(ip) {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\" due to whitelist", nil).Error())
				}
				return
			}
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Handling config request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			err = broker.handleConfigConnection(node.GetName(), netConn)
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle config request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
				bytesSend, err := Tcp.Send(netConn, Message.NewAsync("error", node.GetName(), Error.New("failed to handle config request", err).Error()).Serialize(), broker.application.GetBrokerComponentConfig().TcpTimeoutMs)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send error response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
					}
				} else {
					broker.bytesSentCounter.Add(bytesSend)
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled config request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
				}
				bytesSend, err := Tcp.Send(netConn, Message.NewAsync("success", node.GetName(), "").Serialize(), broker.application.GetBrokerComponentConfig().TcpTimeoutMs)
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send success response to config connection \""+netConn.RemoteAddr().String()+"\"", err).Error())
					}
				} else {
					broker.bytesSentCounter.Add(bytesSend)
				}
			}
			netConn.Close()
		}()
	}
}

func (broker *brokerComponent) handleConfigConnection(nodeName string, netConn net.Conn) error {
	messageBytes, bytesSent, err := Tcp.Receive(netConn, broker.application.GetBrokerComponentConfig().TcpTimeoutMs, broker.application.GetBrokerComponentConfig().IncomingMessageByteLimit)
	if err != nil {
		return Error.New("Failed to receive connection request", err)
	}
	broker.bytesReceivedCounter.Add(bytesSent)
	message := Message.Deserialize(messageBytes)
	if message == nil {
		return Error.New("Invalid connection request", nil)
	}
	err = broker.validateMessage(message)
	if err != nil {
		return Error.New("Invalid connection request message", err)
	}
	err = broker.handleConfigRequest(nodeName, message)
	if err != nil {
		return Error.New("Failed to handle config request with topic \""+message.GetTopic()+"\"", err)
	}
	return nil
}

func (broker *brokerComponent) handleConfigRequest(nodeName string, message *Message.Message) error {
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
		err := broker.addResolverTopicsRemotely(nodeName, payloadSegments...)
		if err != nil {
			return Error.New("Failed to add topics remotely", err)
		}
	case "removeSyncTopics":
		broker.removeSyncTopics(payloadSegments...)
		err := broker.removeResolverTopicsRemotely(nodeName, payloadSegments...)
		if err != nil {
			return Error.New("Failed to remove topics remotely", err)
		}
	case "addAsyncTopics":
		broker.addAsyncTopics(payloadSegments...)
		err := broker.addResolverTopicsRemotely(nodeName, payloadSegments...)
		if err != nil {
			return Error.New("Failed to add topics remotely", err)
		}
	case "removeAsyncTopics":
		broker.removeAsyncTopics(payloadSegments...)
		err := broker.removeResolverTopicsRemotely(nodeName, payloadSegments...)
		if err != nil {
			return Error.New("Failed to remove topics remotely", err)
		}
	default:
		return Error.New("Unknown topic \""+message.GetTopic()+"\"", nil)
	}
	return nil
}
