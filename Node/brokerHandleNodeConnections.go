package Node

import (
	"net"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (node *Node) handleBrokerNodeConnections() {
	broker_ := node.broker
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handling broker node connections", nil).Error())
	}
	for {
		broker := node.broker
		if broker != broker_ {
			return
		}
		netConn, err := broker.brokerTcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept connection request", err).Error())
			}
			return
		}
		go func() {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Accepted connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
			if broker.brokerTcpServer.GetBlacklist().Contains(ip) {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
				}
				return
			}
			if broker.brokerTcpServer.GetWhitelist().ElementCount() > 0 && !broker.brokerTcpServer.GetWhitelist().Contains(ip) {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
				}
				return
			}
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Handling connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			nodeConnection, err := broker.handleNodeConnection(node.GetName(), netConn)
			if err != nil {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle connection request from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled connection request from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
				}
				go node.handleNodeConnectionMessages(nodeConnection)
			}
		}()
	}
}

func (broker *brokerComponent) handleNodeConnection(nodeName string, netConn net.Conn) (*nodeConnection, error) {
	messageBytes, bytesReceived, err := Tcp.Receive(netConn, broker.application.GetBrokerComponentConfig().TcpTimeoutMs, broker.application.GetBrokerComponentConfig().IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive connection request", err)
	}
	broker.bytesReceivedCounter.Add(bytesReceived)
	message := Message.Deserialize(messageBytes)
	if message == nil || message.GetTopic() != "connect" || message.GetOrigin() == "" || message.GetPayload() != "" {
		return nil, Error.New("Invalid connection request \""+string(messageBytes)+"\"", nil)
	}
	err = broker.validateMessage(message)
	if err != nil {
		return nil, Error.New("Invalid connection request message", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync("connected", nodeName, "").Serialize(), broker.application.GetBrokerComponentConfig().TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send message", err)
	}
	nodeConnection := broker.newNodeConnection(message.GetOrigin(), netConn)
	err = broker.addNodeConnection(nodeConnection)
	if err != nil {
		return nil, Error.New("Failed to add node \""+nodeConnection.name+"\"", err)
	}
	broker.bytesSentCounter.Add(bytesSent)
	return nodeConnection, nil
}
