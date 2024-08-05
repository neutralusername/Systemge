package Node

import (
	"net"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (node *Node) handleIncomingConnections() {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting incoming connection handler", nil).Error())
	}
	systemge_ := node.systemge
	for {
		systemge := node.systemge
		if systemge == nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting incoming connection handler because systemge is nil likely due to node being stopped", nil).Error())
			}
			return
		}
		if systemge != systemge_ {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting incoming connection handler because systemge has changed likely due to node restart", nil).Error())
			}
			return
		}
		netConn, err := systemge.tcpServer.GetListener().Accept()
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to accept incoming connection", err).Error())
			}
			continue
		}
		systemge.incomingConnectionAttempts.Add(1)
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Handling incoming connection from "+netConn.RemoteAddr().String(), nil).Error())
		}
		go func() {
			address := netConn.RemoteAddr().String()
			ip, _, err := net.SplitHostPort(address)
			if err != nil {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to split host and port from address \""+address+"\"", err).Error())
				}
				return
			}
			if systemge.tcpServer.GetBlacklist().Contains(ip) {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected incoming connection from \""+netConn.RemoteAddr().String()+"\" due to blacklist", nil).Error())
				}
				return
			}
			if systemge.tcpServer.GetWhitelist().ElementCount() > 0 && !systemge.tcpServer.GetWhitelist().Contains(ip) {
				netConn.Close()
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Rejected incoming connection from \""+netConn.RemoteAddr().String()+"\" due to whitelist", nil).Error())
				}
				return
			}
			incomingConnection, err := systemge.handleIncomingConnection(node.GetName(), netConn)
			if err != nil {
				systemge.incomingConnectionAttemptsFailed.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle incoming connection from "+netConn.RemoteAddr().String(), err).Error())
				}
			} else {
				err := systemge.addIncomingConnection(incomingConnection)
				if err != nil {
					systemge.incomingConnectionAttemptsFailed.Add(1)
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to add incoming connection from "+netConn.RemoteAddr().String()+" with name \""+incomingConnection.name+"\"", err).Error())
					}
					return
				}
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Successfully handled incoming connection from "+netConn.RemoteAddr().String()+" with name \""+incomingConnection.name+"\"", nil).Error())
				}
				systemge.incomingConnectionAttemptsSuccessful.Add(1)
				go node.handleIncomingConnectionMessages(incomingConnection)
			}
		}()
	}
}

func (systemge *systemgeComponent) handleIncomingConnection(nodeName string, netConn net.Conn) (*incomingConnection, error) {
	systemge.incomingConnectionAttempts.Add(1)
	incomingConnection := incomingConnection{
		netConn: netConn,
	}
	messageBytes, err := incomingConnection.assembleMessage(systemge.config.TcpBufferBytes)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+connection_nodeName_topic+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.incomingConnectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+connection_nodeName_topic+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+connection_nodeName_topic+"\" message", err)
	}
	if message.GetTopic() != connection_nodeName_topic {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+connection_nodeName_topic+"\"", nil)
	}
	incomingConnectionName := message.GetPayload()
	if incomingConnectionName == "" {
		netConn.Close()
		return nil, Error.New("Received empty payload in \""+connection_nodeName_topic+"\" message", nil)
	}
	if systemge.config.MaxNodeNameSize != 0 && len(incomingConnectionName) > int(systemge.config.MaxNodeNameSize) {
		netConn.Close()
		return nil, Error.New("Received node name \""+incomingConnectionName+"\" exceeds maximum size of "+Helpers.Uint64ToString(systemge.config.MaxNodeNameSize), nil)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(connection_nodeName_topic, nodeName).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+connection_nodeName_topic+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.incomingConnectionAttemptBytesSent.Add(bytesSent)
	responsibleTopics := []string{}
	systemge.syncMessageHandlerMutex.Lock()
	for topic := range systemge.application.GetSyncMessageHandlers() {
		responsibleTopics = append(responsibleTopics, topic)
	}
	systemge.syncMessageHandlerMutex.Unlock()
	systemge.asyncMessageHandlerMutex.Lock()
	for topic := range systemge.application.GetAsyncMessageHandlers() {
		responsibleTopics = append(responsibleTopics, topic)
	}
	systemge.asyncMessageHandlerMutex.Unlock()
	bytesSent, err = Tcp.Send(netConn, Message.NewAsync(connection_responsibleTopics_topic, Helpers.JsonMarshal(responsibleTopics)).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+connection_responsibleTopics_topic+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.incomingConnectionAttemptBytesSent.Add(bytesSent)
	incomingConn := systemge.newIncomingConnection(netConn, incomingConnectionName)
	incomingConn.tcpBuffer = incomingConnection.tcpBuffer
	return &incomingConnection, nil
}
