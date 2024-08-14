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

type outgoingConnectionAttempt struct {
	endpointConfig *Config.TcpEndpoint
	attempts       uint64

	isAborted bool
}

// repeatedly attempts to establish a connection to an endpoint until either:
// the connection is established,
// the maximum number of attempts is reached,
// the attempt is aborted
// or the systemge component is stopped
func (systemge *SystemgeClient) attemptOutgoingConnection(endpointConfig *Config.TcpEndpoint, transient bool) error {
	address, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return Error.New("Failed to normalize address \""+endpointConfig.Address+"\"", err)
	}
	endpointConfig.Address = address

	systemge.outgoingConnectionMutex.Lock()
	if systemge.outgoingConnections[endpointConfig.Address] != nil {
		systemge.outgoingConnectionMutex.Unlock()
		return Error.New("Outgoing connection with endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	if systemge.outgoingConnectionAttempts[endpointConfig.Address] != nil {
		systemge.outgoingConnectionMutex.Unlock()
		return Error.New("Outgoing connection attempt to endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	outgoingConnectionAttempt := &outgoingConnectionAttempt{
		endpointConfig: endpointConfig,
	}
	systemge.outgoingConnectionAttempts[endpointConfig.Address] = outgoingConnectionAttempt
	systemge.outgoingConnectionMutex.Unlock()

	defer func() {
		outgoingConnectionAttempt.isAborted = true
		systemge.outgoingConnectionMutex.Lock()
		delete(systemge.outgoingConnectionAttempts, outgoingConnectionAttempt.endpointConfig.Address)
		systemge.outgoingConnectionMutex.Unlock()
	}()

	if infoLogger := systemge.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	for {
		select {
		case <-systemge.stopChannel:
			return Error.New("Outgoing connection attempt to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" was stopped because systemge was stopped", nil)
		default:
			if maxConnectionAttempts := systemge.config.MaxConnectionAttempts; maxConnectionAttempts > 0 && outgoingConnectionAttempt.attempts >= maxConnectionAttempts {
				return Error.New("Max attempts reached to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", nil)
			}
			if outgoingConnectionAttempt.isAborted {
				return Error.New("Outgoing connection attempt to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" was aborted", nil)
			}
			if outgoingConnectionAttempt.attempts > 0 {
				time.Sleep(time.Duration(systemge.config.ConnectionAttemptDelayMs) * time.Millisecond)
			}
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", nil).Error())
			}
			outgoingConn, err := systemge.handleOutgoingConnectionHandshake(outgoingConnectionAttempt)
			if err != nil {
				systemge.connectionAttemptsFailed.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", err).Error())
				}
				outgoingConnectionAttempt.attempts++
				continue
			}

			systemge.outgoingConnectionMutex.Lock()
			if outgoingConnectionAttempt.isAborted {
				systemge.outgoingConnectionMutex.Unlock()
				systemge.connectionAttemptsFailed.Add(1)
				return Error.New("Failed to find outgoing connection attempt for endpoint \""+outgoingConn.endpointConfig.Address+"\"", nil)
			}
			if systemge.outgoingConnections[outgoingConn.endpointConfig.Address] != nil {
				systemge.outgoingConnectionMutex.Unlock()
				systemge.connectionAttemptsFailed.Add(1)
				return Error.New("Outgoing connection with endpoint \""+outgoingConn.endpointConfig.Address+"\" already exists", nil)
			}
			systemge.outgoingConnections[outgoingConn.endpointConfig.Address] = outgoingConn
			for topic := range outgoingConn.topics {
				topicResolutions := systemge.topicResolutions[topic]
				if topicResolutions == nil {
					topicResolutions = make(map[string]*outgoingConnection)
					systemge.topicResolutions[topic] = topicResolutions
				}
				topicResolutions[outgoingConn.name] = outgoingConn
			}
			outgoingConn.isTransient = transient
			systemge.outgoingConnectionMutex.Unlock()

			systemge.connectionAttemptsSuccessful.Add(1)
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" with name \""+outgoingConn.name+"\"", nil).Error())
			}
			go systemge.handleOutgoingConnectionMessages(outgoingConn)
			return nil
		}
	}
}

func (systemge *SystemgeClient) handleOutgoingConnectionHandshake(outgoingConnectionAttempt *outgoingConnectionAttempt) (*outgoingConnection, error) {
	systemge.connectionAttempts.Add(1)
	netConn, err := Tcp.NewClient(outgoingConnectionAttempt.endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NODENAME, systemge.config.Name).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.connectionAttemptBytesSent.Add(bytesSent)
	outgoingConnection := outgoingConnection{
		netConn: netConn,
	}
	messageBytes, err := outgoingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NODENAME+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
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
	if systemge.config.MaxNodeNameSize != 0 && len(endpointName) > int(systemge.config.MaxNodeNameSize) {
		netConn.Close()
		return nil, Error.New("Received node name \""+endpointName+"\" exceeds maximum size of "+Helpers.Uint64ToString(systemge.config.MaxNodeNameSize), nil)
	}
	messageBytes, err = outgoingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+Message.TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.connectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err = Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
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
	outgoingConn := systemge.newOutgoingConnection(netConn, outgoingConnectionAttempt.endpointConfig, endpointName, topicMap)
	outgoingConn.tcpBuffer = outgoingConnection.tcpBuffer
	return outgoingConn, nil
}
