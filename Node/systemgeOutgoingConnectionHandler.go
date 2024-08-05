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

func (node *Node) StartOutgoingConnectionLoop(endpointConfig *Config.TcpEndpoint) {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	connectionAttempts := uint64(0)
	systemge_ := node.systemge
	if systemge_ == nil {
		if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
			warningLogger.Log(Error.New("Aborting connection attempts to endpoint \""+endpointConfig.Address+"\" because systemge is nil likely due to node being stopped", nil).Error())
		}
		return
	}
	systemge_.outgoingConnectionMutex.Lock()
	if systemge_.currentlyInOutgoingConnectionLoop[endpointConfig.Address] != nil {
		if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
			warningLogger.Log(Error.New("Aborting connection attempts to endpoint \""+endpointConfig.Address+"\" because loop is already ongoing", nil).Error())
		}
		systemge_.outgoingConnectionMutex.Unlock()
		return
	}
	loopOngoing := true
	systemge_.currentlyInOutgoingConnectionLoop[endpointConfig.Address] = &loopOngoing
	systemge_.outgoingConnectionMutex.Unlock()
	defer func() {
		systemge_.outgoingConnectionMutex.Lock()
		delete(systemge_.currentlyInOutgoingConnectionLoop, endpointConfig.Address)
		systemge_.outgoingConnectionMutex.Unlock()
	}()
	for {
		systemge := node.systemge
		if systemge == nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting connection attempts to endpoint \""+endpointConfig.Address+"\" because systemge is nil likely due to node being stopped", nil).Error())
			}
			return
		}
		if systemge != systemge_ {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting connection attempts to endpoint \""+endpointConfig.Address+"\" because systemge has changed likely due to node restart", nil).Error())
			}
			return
		}
		if !loopOngoing {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Aborting connection attempts to endpoint \""+endpointConfig.Address+"\" because loop was cancelled", nil).Error())
			}
			return
		}
		if maxConnectionAttempts := systemge.config.MaxConnectionAttempts; maxConnectionAttempts > 0 && connectionAttempts >= maxConnectionAttempts {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Max attempts reached to connect to endpoint \""+endpointConfig.Address+"\"", nil).Error())
			}
			if systemge.config.StopAfterOutgoingConnectionLoss {
				if err := node.stop(true); err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to stop node", err).Error())
					}
				}
			}
			return
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(connectionAttempts)+" to connect to endpoint \""+endpointConfig.Address+"\"", nil).Error())
		}
		if nodeConnection, err := systemge.connectionAttempt(node.GetName(), endpointConfig); err != nil {
			systemge.outgoingConnectionAttemptsFailed.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(connectionAttempts)+" to connect to endpoint \""+endpointConfig.Address+"\"", err).Error())
			}
			connectionAttempts++
			time.Sleep(time.Duration(systemge.config.ConnectionAttemptDelayMs) * time.Millisecond)
		} else {
			err := systemge.addOutgoingConnection(nodeConnection)
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add outgoing connection", err).Error())
				}
				return
			}
			systemge.outgoingConnectionAttemptsSuccessful.Add(1)
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(connectionAttempts)+" to connect to endpoint \""+endpointConfig.Address+"\" with name \""+nodeConnection.name+"\"", nil).Error())
			}
			go node.handleOutgoingConnectionMessages(nodeConnection)
			return
		}
	}
}

func (systemge *systemgeComponent) connectionAttempt(nodeName string, endpointConfig *Config.TcpEndpoint) (*outgoingConnection, error) {
	systemge.outgoingConnectionAttempts.Add(1)
	netConn, err := Tcp.NewClient(endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to endpoint \""+endpointConfig.Address+"\"", err)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(connection_nodeName_topic, nodeName).Serialize(), systemge.config.TcpTimeoutMs)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to send \""+connection_nodeName_topic+"\" message", err)
	}
	systemge.bytesSent.Add(bytesSent)
	systemge.outgoingConnectionAttemptBytesSent.Add(bytesSent)
	outgoingConnection := outgoingConnection{
		netConn: netConn,
	}
	messageBytes, err := outgoingConnection.assembleMessage(systemge.config.TcpBufferBytes)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+connection_nodeName_topic+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.outgoingConnectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
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
	endpointName := message.GetPayload()
	if endpointName == "" {
		netConn.Close()
		return nil, Error.New("Received empty payload in \""+connection_nodeName_topic+"\" message", nil)
	}
	if systemge.config.MaxNodeNameSize != 0 && len(endpointName) > int(systemge.config.MaxNodeNameSize) {
		netConn.Close()
		return nil, Error.New("Received node name \""+endpointName+"\" exceeds maximum size of "+Helpers.Uint64ToString(systemge.config.MaxNodeNameSize), nil)
	}
	messageBytes, err = outgoingConnection.assembleMessage(systemge.config.TcpBufferBytes)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+connection_responsibleTopics_topic+"\" message", err)
	}
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	systemge.outgoingConnectionAttemptBytesReceived.Add(uint64(len(messageBytes)))
	message, err = Message.Deserialize(messageBytes, "")
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to deserialize \""+connection_responsibleTopics_topic+"\" message", err)
	}
	if err := systemge.validateMessage(message); err != nil {
		netConn.Close()
		return nil, Error.New("Failed to validate \""+connection_responsibleTopics_topic+"\" message", err)
	}
	if message.GetTopic() != connection_responsibleTopics_topic {
		netConn.Close()
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+connection_responsibleTopics_topic+"\"", nil)
	}
	topics := []string{}
	json.Unmarshal([]byte(message.GetPayload()), &topics)
	outgoingConn := systemge.newOutgoingConnection(netConn, endpointConfig, endpointName, topics)
	outgoingConn.tcpBuffer = outgoingConnection.tcpBuffer
	return outgoingConn, nil
}
