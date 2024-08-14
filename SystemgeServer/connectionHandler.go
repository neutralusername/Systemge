package Node

import (
	"net"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

// handles incoming connections from other nodes one at a time until systemge is stopped
func (systemge *SystemgeServer) handleIncomingConnections() {
	if infoLogger := systemge.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting incoming connection handler", nil).Error())
	}
	for {
		select {
		case <-systemge.stopChannel:
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Stopping incoming connection handler for node-session-id \""+Helpers.GetPointerId(systemge.stopChannel)+"\"", nil).Error())
			}
			close(systemge.incomingConnectionsStopChannel)
			return
		default:
			netConn, err := systemge.tcpServer.GetListener().Accept()
			if err != nil {
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to accept incoming connection", err).Error())
				}
				continue
			}
			systemge.connectionAttempts.Add(1)
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handling incoming connection from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			if err := systemge.accessControlIncomingConnection(netConn); err != nil {
				systemge.connectionAttemptsFailed.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log("Rejected incoming connection from \"" + netConn.RemoteAddr().String() + "\" due to access control: " + err.Error())
				}
				netConn.Close()
				continue
			}
			incomingConnection, err := systemge.handleIncomingConnectionHandshake(netConn)
			if err != nil {
				systemge.connectionAttemptsFailed.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle incoming connection handshake from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
				netConn.Close()
				continue
			}

			systemge.incomingConnectionMutex.Lock()
			if systemge.incomingConnections[incomingConnection.name] != nil {
				systemge.incomingConnectionMutex.Unlock()
				systemge.connectionAttemptsFailed.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add incoming connection from \""+netConn.RemoteAddr().String()+"\" with name \""+incomingConnection.name+"\"", err).Error())
				}
				netConn.Close()
				continue
			}
			systemge.incomingConnections[incomingConnection.name] = incomingConnection
			systemge.incomingConnectionMutex.Unlock()

			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Successfully handled incoming connection from \""+netConn.RemoteAddr().String()+"\" with name \""+incomingConnection.name+"\"", nil).Error())
			}
			systemge.connectionAttemptsSuccessful.Add(1)
			go systemge.handleIncomingConnectionMessages(incomingConnection)
		}
	}
}

func (systemge *SystemgeServer) handleIncomingConnectionHandshake(netConn net.Conn) (*incomingConnection, error) {
	systemge.connectionAttempts.Add(1)
	incomingConnection := incomingConnection{
		netConn: netConn,
	}
	messageBytes, err := incomingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive \""+TOPIC_NODENAME+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+TOPIC_NODENAME+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
		return nil, Error.New("Failed to validate \""+TOPIC_NODENAME+"\" message", err)
	}
	if message.GetTopic() != TOPIC_NODENAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+TOPIC_NODENAME+"\"", nil)
	}
	incomingConnectionName := message.GetPayload()
	if incomingConnectionName == "" {
		return nil, Error.New("Received empty payload in \""+TOPIC_NODENAME+"\" message", nil)
	}
	if systemge.config.MaxNodeNameSize != 0 && len(incomingConnectionName) > int(systemge.config.MaxNodeNameSize) {
		return nil, Error.New("Received node name \""+incomingConnectionName+"\" exceeds maximum size of "+Helpers.Uint64ToString(systemge.config.MaxNodeNameSize), nil)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(TOPIC_NODENAME, systemge.config.Name).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+TOPIC_NODENAME+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.connectionAttemptBytesSent.Add(bytesSent)
	responsibleTopics := []string{}
	systemge.syncMessageHandlerMutex.RLock()
	for topic := range systemge.syncMessageHandlers {
		responsibleTopics = append(responsibleTopics, topic)
	}
	systemge.syncMessageHandlerMutex.RUnlock()
	systemge.asyncMessageHandlerMutex.RLock()
	for topic := range systemge.asyncMessageHandlers {
		responsibleTopics = append(responsibleTopics, topic)
	}
	systemge.asyncMessageHandlerMutex.RUnlock()
	bytesSent, err = Tcp.Send(netConn, Message.NewAsync(TOPIC_RESPONSIBLETOPICS, Helpers.JsonMarshal(responsibleTopics)).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.connectionAttemptBytesSent.Add(bytesSent)
	incomingConn := systemge.newIncomingConnection(netConn, incomingConnectionName)
	incomingConn.tcpBuffer = incomingConnection.tcpBuffer
	return incomingConn, nil
}

func (systemge *SystemgeServer) accessControlIncomingConnection(netConn net.Conn) error {
	address := netConn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to split host and port from address", err)
	}
	if systemge.config.IpRateLimiter != nil && !systemge.ipRateLimiter.RegisterConnectionAttempt(ip) {
		return Error.New("Rejected incoming connection from \""+address+"\" due to ip rate limiting", nil)
	}
	if systemge.tcpServer.GetBlacklist().Contains(ip) {
		return Error.New("Rejected incoming connection due to blacklist", nil)
	}
	if systemge.tcpServer.GetWhitelist().ElementCount() > 0 && !systemge.tcpServer.GetWhitelist().Contains(ip) {
		return Error.New("Rejected incoming connection due to whitelist", nil)
	}
	return nil
}
