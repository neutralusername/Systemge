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

	stopChannel chan bool //closing of this channel indicates that the outgoing connection attempt has finished its ongoing tasks.
}

// repeatedly attempts to establish a connection to an endpoint until either:
// the connection is established,
// the maximum number of attempts is reached,
// the attempt is aborted
// or the systemge component is stopped
func (systemge *systemgeComponent) attemptOutgoingConnection(endpointConfig *Config.TcpEndpoint, transient bool) error {
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
		stopChannel:    make(chan bool),
	}
	systemge.outgoingConnectionAttempts[endpointConfig.Address] = outgoingConnectionAttempt
	systemge.outgoingConnectionMutex.Unlock()

	defer func() {
		outgoingConnectionAttempt.isAborted = true
		close(outgoingConnectionAttempt.stopChannel)
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
				systemge.outgoingConnectionAttemptsFailed.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", err).Error())
				}
				outgoingConnectionAttempt.attempts++
				continue
			}

			// in the current version there is a potential deadlock caused by these locks. there are also various race conditions caused by stopping during startup.
			// just as a heads up until the fixes are done, be careful when stopping nodes during startup. (it works most of the time, especially with some pauses in between)
			systemge.outgoingConnectionMutex.Lock()
			if outgoingConnectionAttempt.isAborted {
				systemge.outgoingConnectionMutex.Unlock()
				systemge.outgoingConnectionAttemptsFailed.Add(1)
				return Error.New("Failed to find outgoing connection attempt for endpoint \""+outgoingConn.endpointConfig.Address+"\"", nil)
			}
			if systemge.outgoingConnections[outgoingConn.endpointConfig.Address] != nil {
				systemge.outgoingConnectionMutex.Unlock()
				systemge.outgoingConnectionAttemptsFailed.Add(1)
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

			systemge.outgoingConnectionAttemptsSuccessful.Add(1)
			if infoLogger := systemge.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(outgoingConnectionAttempt.attempts)+" to connect to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\" with name \""+outgoingConn.name+"\"", nil).Error())
			}
			go systemge.handleOutgoingConnectionMessages(outgoingConn)
			return nil
		}
	}
}

func (systemge *systemgeComponent) handleOutgoingConnectionHandshake(outgoingConnectionAttempt *outgoingConnectionAttempt) (*outgoingConnection, error) {
	systemge.outgoingConnectionAttemptsCount.Add(1)
	netConn, err := Tcp.NewClient(outgoingConnectionAttempt.endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to endpoint \""+outgoingConnectionAttempt.endpointConfig.Address+"\"", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(TOPIC_NODENAME, systemge.nodeName).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+TOPIC_NODENAME+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.outgoingConnectionAttemptBytesSent.Add(bytesSent)
	outgoingConnection := outgoingConnection{
		netConn: netConn,
	}
	messageBytes, err := outgoingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+TOPIC_NODENAME+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.outgoingConnectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+TOPIC_NODENAME+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+TOPIC_NODENAME+"\" message", err)
	}
	if message.GetTopic() != TOPIC_NODENAME {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+TOPIC_NODENAME+"\"", nil)
	}
	endpointName := message.GetPayload()
	if endpointName == "" {
		netConn.Close()
		return nil, Error.New("Received empty payload in \""+TOPIC_NODENAME+"\" message", nil)
	}
	if systemge.config.MaxNodeNameSize != 0 && len(endpointName) > int(systemge.config.MaxNodeNameSize) {
		netConn.Close()
		return nil, Error.New("Received node name \""+endpointName+"\" exceeds maximum size of "+Helpers.Uint64ToString(systemge.config.MaxNodeNameSize), nil)
	}
	messageBytes, err = outgoingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.outgoingConnectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err = Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+TOPIC_RESPONSIBLETOPICS+"\" message", err)
	}
	if message.GetTopic() != TOPIC_RESPONSIBLETOPICS {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+TOPIC_RESPONSIBLETOPICS+"\"", nil)
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
