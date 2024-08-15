package SystemgeServer

import (
	"net"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

// handles client connections, one at a time until systemge is stopped
func (server *SystemgeServer) handleClientConnections() {
	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting client connection handler", nil).Error())
	}
	for {
		select {
		case <-server.stopChannel:
			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Stopping client connection handler", nil).Error())
			}
			close(server.clientConnectionStopChannelStopChannel)
			return
		default:
			netConn, err := server.tcpServer.GetListener().Accept()
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to accept client connection", err).Error())
				}
				continue
			}
			server.connectionAttempts.Add(1)
			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Handling client connection from \""+netConn.RemoteAddr().String()+"\"", nil).Error())
			}
			if err := server.accessControlClientConnection(netConn); err != nil {
				server.connectionAttemptsFailed.Add(1)
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log("Rejected client connection from \"" + netConn.RemoteAddr().String() + "\" due to access control: " + err.Error())
				}
				netConn.Close()
				continue
			}
			clientConnection, err := server.handleClientConnectionHandshake(netConn)
			if err != nil {
				server.connectionAttemptsFailed.Add(1)
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle client connection handshake from \""+netConn.RemoteAddr().String()+"\"", err).Error())
				}
				netConn.Close()
				continue
			}

			server.clientConnectionMutex.Lock()
			if server.clientConnections[clientConnection.name] != nil {
				server.clientConnectionMutex.Unlock()
				server.connectionAttemptsFailed.Add(1)
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add client connection from \""+netConn.RemoteAddr().String()+"\" with name \""+clientConnection.name+"\"", err).Error())
				}
				netConn.Close()
				continue
			}
			server.clientConnections[clientConnection.name] = clientConnection
			server.clientConnectionMutex.Unlock()

			if infoLogger := server.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Successfully handled client connection from \""+netConn.RemoteAddr().String()+"\" with name \""+clientConnection.name+"\"", nil).Error())
			}
			server.connectionAttemptsSuccessful.Add(1)
			go server.handleClientConnectionMessages(clientConnection)
		}
	}
}

func (server *SystemgeServer) handleClientConnectionHandshake(netConn net.Conn) (*clientConnection, error) {
	server.connectionAttempts.Add(1)
	clientConnection := &clientConnection{
		netConn: netConn,
	}
	messageBytes, err := clientConnection.receiveMessage(server.config.TcpBufferBytes, server.config.ClientMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	server.bytesReceived.Add(uint64(len(messageBytes)))
	server.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if err := server.validateMessage(message); err != nil {
		return nil, Error.New("Failed to validate \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	clientConnectionName := message.GetPayload()
	if clientConnectionName == "" {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	if server.config.MaxClientNameLength != 0 && len(clientConnectionName) > int(server.config.MaxClientNameLength) {
		return nil, Error.New("Received client name \""+clientConnectionName+"\" exceeds maximum size of "+Helpers.Uint64ToString(server.config.MaxClientNameLength), nil)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, server.config.Name).Serialize(), server.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	server.bytesSent.Add(bytesSent)
	server.connectionAttemptBytesSent.Add(bytesSent)

	tcpBuffer := clientConnection.tcpBuffer
	clientConnection = server.newClientConnection(netConn, clientConnectionName)
	clientConnection.tcpBuffer = tcpBuffer
	return clientConnection, nil
}

func (systemge *SystemgeServer) accessControlClientConnection(netConn net.Conn) error {
	address := netConn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to split host and port from address", err)
	}
	if systemge.config.IpRateLimiter != nil && !systemge.ipRateLimiter.RegisterConnectionAttempt(ip) {
		return Error.New("Rejected client connection from \""+address+"\" due to ip rate limiting", nil)
	}
	if systemge.tcpServer.GetBlacklist().Contains(ip) {
		return Error.New("Rejected client connection due to blacklist", nil)
	}
	if systemge.tcpServer.GetWhitelist().ElementCount() > 0 && !systemge.tcpServer.GetWhitelist().Contains(ip) {
		return Error.New("Rejected client connection due to whitelist", nil)
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
