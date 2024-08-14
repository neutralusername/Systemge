package Node

import (
	"net"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

// handles incoming connections from other nodes one at a time until systemge is stopped
func (server *SystemgeServer) handleIncomingConnections() {
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting incoming connection handler", nil).Error())
	}
	for {
		select {
		case <-server.stopChannel:
			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Stopping incoming connection handler for node-session-id \""+Helpers.GetPointerId(server.stopChannel)+"\"", nil).Error())
			}
			close(server.incomingConnectionsStopChannel)
			return
		default:
			netConn, err := server.tcpServer.GetListener().Accept()
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to accept incoming connection", err).Error())
				}
				continue
			}
			server.connectionAttempts.Add(1)
			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handling incoming connection from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			if err := server.accessControlIncomingConnection(netConn); err != nil {
				server.connectionAttemptsFailed.Add(1)
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log("Rejected incoming connection from \"" + netConn.RemoteAddr().String() + "\" due to access control: " + err.Error())
				}
				netConn.Close()
				continue
			}
			incomingConnection, err := server.handleIncomingConnectionHandshake(netConn)
			if err != nil {
				server.connectionAttemptsFailed.Add(1)
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle incoming connection handshake from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
				netConn.Close()
				continue
			}

			server.clientConnectionMutex.Lock()
			if server.clientConnections[incomingConnection.name] != nil {
				server.clientConnectionMutex.Unlock()
				server.connectionAttemptsFailed.Add(1)
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add incoming connection from \""+netConn.RemoteAddr().String()+"\" with name \""+incomingConnection.name+"\"", err).Error())
				}
				netConn.Close()
				continue
			}
			server.clientConnections[incomingConnection.name] = incomingConnection
			server.clientConnectionMutex.Unlock()

			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Successfully handled incoming connection from \""+netConn.RemoteAddr().String()+"\" with name \""+incomingConnection.name+"\"", nil).Error())
			}
			server.connectionAttemptsSuccessful.Add(1)
			go server.handleIncomingConnectionMessages(incomingConnection)
		}
	}
}

func (server *SystemgeServer) handleIncomingConnectionHandshake(netConn net.Conn) (*clientConnection, error) {
	server.connectionAttempts.Add(1)
	incomingConnection := clientConnection{
		netConn: netConn,
	}
	messageBytes, err := incomingConnection.receiveMessage(server.config.TcpBufferBytes, server.config.IncomingMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	server.bytesReceived.Add(uint64(len(messageBytes)))
	server.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	if err := server.validateMessage(message); err != nil {
		return nil, Error.New("Failed to validate \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NODENAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NODENAME+"\"", nil)
	}
	clientConnectionName := message.GetPayload()
	if clientConnectionName == "" {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NODENAME+"\" message", nil)
	}
	if server.config.MaxNodeNameSize != 0 && len(clientConnectionName) > int(server.config.MaxNodeNameSize) {
		return nil, Error.New("Received node name \""+clientConnectionName+"\" exceeds maximum size of "+Helpers.Uint64ToString(server.config.MaxNodeNameSize), nil)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NODENAME, server.config.Name).Serialize(), server.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	server.bytesSent.Add(bytesSent)
	server.connectionAttemptBytesSent.Add(bytesSent)
	responsibleTopics := []string{}
	server.syncMessageHandlerMutex.RLock()
	for topic := range server.syncMessageHandlers {
		responsibleTopics = append(responsibleTopics, topic)
	}
	server.syncMessageHandlerMutex.RUnlock()
	server.asyncMessageHandlerMutex.RLock()
	for topic := range server.asyncMessageHandlers {
		responsibleTopics = append(responsibleTopics, topic)
	}
	server.asyncMessageHandlerMutex.RUnlock()
	bytesSent, err = Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_RESPONSIBLETOPICS, Helpers.JsonMarshal(responsibleTopics)).Serialize(), server.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	server.bytesSent.Add(bytesSent)
	server.connectionAttemptBytesSent.Add(bytesSent)
	clientConnection := server.newClientConnection(netConn, clientConnectionName)
	clientConnection.tcpBuffer = incomingConnection.tcpBuffer
	return clientConnection, nil
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

func (systemge *SystemgeServer) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := systemge.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncTokenToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := systemge.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := systemge.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
