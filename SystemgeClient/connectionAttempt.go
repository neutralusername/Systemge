package SystemgeClient

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

type serverConnectionAttempt struct {
	endpointConfig *Config.TcpEndpoint
	attempts       uint64

	isAborted bool
}

// repeatedly attempts to establish a connection to an endpoint until either:
// the connection is established,
// the maximum number of attempts is reached,
// the attempt is aborted
// or the systemge component is stopped
func (client *SystemgeClient) attemptServerConnection(endpointConfig *Config.TcpEndpoint, transient bool) error {
	address, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return Error.New("Failed to normalize address \""+endpointConfig.Address+"\"", err)
	}
	endpointConfig.Address = address

	client.serverConnectionMutex.Lock()
	if client.serverConnections[endpointConfig.Address] != nil {
		client.serverConnectionMutex.Unlock()
		return Error.New("server connection with endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	if client.serverConnectionAttempts[endpointConfig.Address] != nil {
		client.serverConnectionMutex.Unlock()
		return Error.New("server connection attempt to endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	serverConnectionAttempt := &serverConnectionAttempt{
		endpointConfig: endpointConfig,
	}
	client.serverConnectionAttempts[endpointConfig.Address] = serverConnectionAttempt
	client.serverConnectionMutex.Unlock()

	defer func() {
		serverConnectionAttempt.isAborted = true
		client.serverConnectionMutex.Lock()
		delete(client.serverConnectionAttempts, serverConnectionAttempt.endpointConfig.Address)
		client.serverConnectionMutex.Unlock()
	}()

	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	for {
		select {
		case <-client.stopChannel:
			return Error.New("server connection attempt to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\" was stopped because systemge was stopped", nil)
		default:
			if maxConnectionAttempts := client.config.MaxConnectionAttempts; maxConnectionAttempts > 0 && serverConnectionAttempt.attempts >= maxConnectionAttempts {
				return Error.New("Max attempts reached to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", nil)
			}
			if serverConnectionAttempt.isAborted {
				return Error.New("server connection attempt to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\" was aborted", nil)
			}
			if serverConnectionAttempt.attempts > 0 {
				time.Sleep(time.Duration(client.config.ConnectionAttemptDelayMs) * time.Millisecond)
			}
			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(serverConnectionAttempt.attempts)+" to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", nil).Error())
			}
			serverConnnection, err := client.handleServerConnectionHandshake(serverConnectionAttempt)
			if err != nil {
				client.connectionAttemptsFailed.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(serverConnectionAttempt.attempts)+" to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", err).Error())
				}
				serverConnectionAttempt.attempts++
				continue
			}

			client.serverConnectionMutex.Lock()
			if serverConnectionAttempt.isAborted {
				client.serverConnectionMutex.Unlock()
				client.connectionAttemptsFailed.Add(1)
				return Error.New("Failed to find server connection attempt for endpoint \""+serverConnnection.endpointConfig.Address+"\"", nil)
			}
			if client.serverConnections[serverConnnection.endpointConfig.Address] != nil {
				client.serverConnectionMutex.Unlock()
				client.connectionAttemptsFailed.Add(1)
				return Error.New("server connection with endpoint \""+serverConnnection.endpointConfig.Address+"\" already exists", nil)
			}
			client.serverConnections[serverConnnection.endpointConfig.Address] = serverConnnection
			for topic := range serverConnnection.topics {
				topicResolutions := client.topicResolutions[topic]
				if topicResolutions == nil {
					topicResolutions = make(map[string]*serverConnection)
					client.topicResolutions[topic] = topicResolutions
				}
				topicResolutions[serverConnnection.name] = serverConnnection
			}
			serverConnnection.isTransient = transient
			client.serverConnectionMutex.Unlock()

			client.connectionAttemptsSuccessful.Add(1)
			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(serverConnectionAttempt.attempts)+" to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\" with name \""+serverConnnection.name+"\"", nil).Error())
			}
			go client.handleServerConnectionMessages(serverConnnection)
			return nil
		}
	}
}

func (client *SystemgeClient) handleServerConnectionHandshake(serverConnectionAttempt *serverConnectionAttempt) (*serverConnection, error) {
	client.connectionAttempts.Add(1)
	netConn, err := Tcp.NewClient(serverConnectionAttempt.endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, client.config.Name).Serialize(), client.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	client.bytesSent.Add(bytesSent)
	client.connectionAttemptBytesSent.Add(bytesSent)
	serverConnection := &serverConnection{
		netConn: netConn,
	}
	messageBytes, err := serverConnection.receiveMessage(client.config.TcpBufferBytes, client.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	client.bytesReceived.Add(uint64(len(messageBytes)))
	client.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if err := client.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	endpointName := message.GetPayload()
	if endpointName == "" {
		netConn.Close()
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	if client.config.MaxServerNameLength != 0 && len(endpointName) > int(client.config.MaxServerNameLength) {
		netConn.Close()
		return nil, Error.New("Received server name \""+endpointName+"\" exceeds maximum size of "+Helpers.Uint64ToString(client.config.MaxServerNameLength), nil)
	}
	messageBytes, err = serverConnection.receiveMessage(client.config.TcpBufferBytes, client.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+Message.TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	client.bytesReceived.Add(uint64(len(messageBytes)))
	client.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err = Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	if err := client.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+Message.TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_RESPONSIBLETOPICS {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_RESPONSIBLETOPICS+"\"", nil)
	}
	topics := []string{}
	json.Unmarshal([]byte(message.GetPayload()), &topics)
	topicMap := map[string]bool{}
	for _, topic := range topics {
		topicMap[topic] = true
	}
	tcpBuffer := serverConnection.tcpBuffer
	serverConnection = client.newServerConnection(netConn, serverConnectionAttempt.endpointConfig, endpointName, topicMap)
	serverConnection.tcpBuffer = tcpBuffer
	return serverConnection, nil
}
