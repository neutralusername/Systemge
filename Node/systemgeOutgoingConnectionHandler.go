package Node

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func (node *Node) outgoingConnectionLoop(endpointConfig *Config.TcpEndpoint) error {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	connectionAttempts := uint64(0)
	systemge_ := node.systemge
	if systemge_ == nil {
		return Error.New("Systemge is nil", nil)
	}
	systemge_.outgoingConnectionMutex.Lock()
	if systemge_.outgoingConnections[endpointConfig.Address] != nil {
		systemge_.outgoingConnectionMutex.Unlock()
		return Error.New("Connection to endpoint \""+endpointConfig.Address+"\" already exists", nil)
	}
	loopOngoing := systemge_.currentlyInOutgoingConnectionLoop[endpointConfig.Address]
	if loopOngoing == nil {
		systemge_.outgoingConnectionMutex.Unlock()
		return Error.New("Connection to endpoint \""+endpointConfig.Address+"\" has been established incorrectly", nil)
	}
	defer func() {
		systemge_.outgoingConnectionMutex.Lock()
		delete(systemge_.currentlyInOutgoingConnectionLoop, endpointConfig.Address)
		systemge_.outgoingConnectionMutex.Unlock()
	}()
	systemge_.outgoingConnectionMutex.Unlock()
	for {
		systemge := node.systemge
		if systemge == nil {
			return Error.New("Systemge is nil", nil)
		}
		if systemge != systemge_ {
			return Error.New("Systemge has changed likely due to node restart", nil)
		}
		if !*loopOngoing {
			return Error.New("Loop was cancelled", nil)
		}
		if maxConnectionAttempts := systemge.config.MaxConnectionAttempts; maxConnectionAttempts > 0 && connectionAttempts >= maxConnectionAttempts {
			if systemge.config.StopAfterOutgoingConnectionLoss {
				if err := node.Stop(); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to stop node", err).Error())
					}
					if mailer := node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to stop node", err).Error()))
						if err != nil {
							if errorLogger := node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log(Error.New("Failed sending mail", err).Error())
							}
						}
					}
				}
			}
			return Error.New("Max attempts reached to connect to endpoint \""+endpointConfig.Address+"\"", nil)
		}
		if connectionAttempts > 0 {
			time.Sleep(time.Duration(systemge.config.ConnectionAttemptDelayMs) * time.Millisecond)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(connectionAttempts)+" to connect to endpoint \""+endpointConfig.Address+"\"", nil).Error())
		}
		if outgoingConnection, err := systemge.connectionAttempt(node.GetName(), endpointConfig); err != nil {
			systemge.outgoingConnectionAttemptsFailed.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(connectionAttempts)+" to connect to endpoint \""+endpointConfig.Address+"\"", err).Error())
			}
			connectionAttempts++
		} else {
			err := systemge.addOutgoingConnection(outgoingConnection)
			if err != nil {
				outgoingConnection.netConn.Close()
				return Error.New("Failed to add outgoing connection", err)
			}
			systemge.outgoingConnectionAttemptsSuccessful.Add(1)
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(connectionAttempts)+" to connect to endpoint \""+endpointConfig.Address+"\" with name \""+outgoingConnection.name+"\"", nil).Error())
			}
			go node.handleOutgoingConnectionMessages(outgoingConnection)
			return nil
		}
	}
}

func (systemge *systemgeComponent) connectionAttempt(nodeName string, endpointConfig *Config.TcpEndpoint) (*outgoingConnection, error) {
	systemge.outgoingConnectionAttempts.Add(1)
	netConn, err := Tcp.NewClient(endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to endpoint \""+endpointConfig.Address+"\"", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(TOPIC_NODENAME, nodeName).Serialize(), systemge.config.TcpTimeoutMs)
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
	outgoingConn := systemge.newOutgoingConnection(netConn, endpointConfig, endpointName, topicMap)
	outgoingConn.tcpBuffer = outgoingConnection.tcpBuffer
	return outgoingConn, nil
}
