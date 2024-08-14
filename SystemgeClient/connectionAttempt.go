package Node

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
func (client *SystemgeClient) attemptOutgoingConnection(endpointConfig *Config.TcpEndpoint, transient bool) error {
	address, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return Error.New("Failed to normalize address \""+endpointConfig.Address+"\"", err)
	}
	endpointConfig.Address = address

	client.serverConnectionMutex.Lock()
	if client.serverConnections[endpointConfig.Address] != nil {
		client.serverConnectionMutex.Unlock()
		return Error.New("Outgoing connection with endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	if client.serverConnectionAttempts[endpointConfig.Address] != nil {
		client.serverConnectionMutex.Unlock()
		return Error.New("Outgoing connection attempt to endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	outgoingConnectionAttempt := &serverConnectionAttempt{
		endpointConfig: endpointConfig,
	}
	client.serverConnectionAttempts[endpointConfig.Address] = outgoingConnectionAttempt
	client.serverConnectionMutex.Unlock()

	defer func() {
		outgoingConnectionAttempt.isAborted = true
		client.serverConnectionMutex.Lock()
		delete(client.serverConnectionAttempts, outgoingConnectionAttempt.endpointConfig.Address)
		client.serverConnectionMutex.Unlock()
	}()

	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	for {
		select {
		case <-client.stopChannel:
			return Error.New("Outgoing connection attempt to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" was stopped because systemge was stopped", nil)
		default:
			if maxConnectionAttempts := client.config.MaxConnectionAttempts; maxConnectionAttempts > 0 && outgoingConnectionAttempt.attempts >= maxConnectionAttempts {
				return Error.New("Max attempts reached to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", nil)
			}
			if outgoingConnectionAttempt.isAborted {
				return Error.New("Outgoing connection attempt to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" was aborted", nil)
			}
			if outgoingConnectionAttempt.attempts > 0 {
				time.Sleep(time.Duration(client.config.ConnectionAttemptDelayMs) * time.Millisecond)
			}
			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", nil).Error())
			}
			outgoingConn, err := client.handleOutgoingConnectionHandshake(outgoingConnectionAttempt)
			if err != nil {
				client.connectionAttemptsFailed.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", err).Error())
				}
				outgoingConnectionAttempt.attempts++
				continue
			}

			client.serverConnectionMutex.Lock()
			if outgoingConnectionAttempt.isAborted {
				client.serverConnectionMutex.Unlock()
				client.connectionAttemptsFailed.Add(1)
				return Error.New("Failed to find outgoing connection attempt for endpoint \""+outgoingConn.endpointConfig.Address+"\"", nil)
			}
			if client.serverConnections[outgoingConn.endpointConfig.Address] != nil {
				client.serverConnectionMutex.Unlock()
				client.connectionAttemptsFailed.Add(1)
				return Error.New("Outgoing connection with endpoint \""+outgoingConn.endpointConfig.Address+"\" already exists", nil)
			}
			client.serverConnections[outgoingConn.endpointConfig.Address] = outgoingConn
			for topic := range outgoingConn.topics {
				topicResolutions := client.topicResolutions[topic]
				if topicResolutions == nil {
					topicResolutions = make(map[string]*serverConnection)
					client.topicResolutions[topic] = topicResolutions
				}
				topicResolutions[outgoingConn.name] = outgoingConn
			}
			outgoingConn.isTransient = transient
			client.serverConnectionMutex.Unlock()

			client.connectionAttemptsSuccessful.Add(1)
			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" with name \""+outgoingConn.name+"\"", nil).Error())
			}
			go client.handleOutgoingConnectionMessages(outgoingConn)
			return nil
		}
	}
}

func (client *SystemgeClient) handleOutgoingConnectionHandshake(outgoingConnectionAttempt *serverConnectionAttempt) (*serverConnection, error) {
	client.connectionAttempts.Add(1)
	netConn, err := Tcp.NewClient(outgoingConnectionAttempt.endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NODENAME, client.config.Name).Serialize(), client.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	client.bytesSent.Add(bytesSent)
	client.connectionAttemptBytesSent.Add(bytesSent)
	outgoingConnection := serverConnection{
		netConn: netConn,
	}
	messageBytes, err := outgoingConnection.receiveMessage(client.config.TcpBufferBytes, client.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	client.bytesReceived.Add(uint64(len(messageBytes)))
	client.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	if err := client.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NODENAME {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NODENAME+"\"", nil)
	}
	endpointName := message.GetPayload()
	if endpointName == "" {
		netConn.Close()
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NODENAME+"\" message", nil)
	}
	if client.config.MaxNodeNameSize != 0 && len(endpointName) > int(client.config.MaxNodeNameSize) {
		netConn.Close()
		return nil, Error.New("Received node name \""+endpointName+"\" exceeds maximum size of "+Helpers.Uint64ToString(client.config.MaxNodeNameSize), nil)
	}
	messageBytes, err = outgoingConnection.receiveMessage(client.config.TcpBufferBytes, client.config.IncomingMessageByteLimit)
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
	outgoingConn := client.newOutgoingConnection(netConn, outgoingConnectionAttempt.endpointConfig, endpointName, topicMap)
	outgoingConn.tcpBuffer = outgoingConnection.tcpBuffer
	return outgoingConn, nil
}
