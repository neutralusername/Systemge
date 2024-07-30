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

func (node *Node) outgoingConnectionLoop(endpointConfig *Config.TcpEndpoint) {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	connectionAttempts := uint64(0)
	systemge_ := node.systemge
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
			infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(connectionAttempts+1)+" to connect to endpoint \""+endpointConfig.Address+"\"", nil).Error())
		}
		if nodeConnection, err := systemge.connectionAttempt(node.GetName(), endpointConfig); err != nil {
			systemge.outgoingConnectionAttemptsFailed.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(connectionAttempts+1)+" to connect to endpoint \""+endpointConfig.Address+"\"", err).Error())
			}
			connectionAttempts++
			time.Sleep(time.Duration(systemge.config.ConnectionAttemptDelayMs) * time.Millisecond)
		} else {
			systemge.outgoingConnectionAttemptsSuccessful.Add(1)
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(connectionAttempts+1)+" to connect to endpoint \""+endpointConfig.Address+"\"", nil).Error())
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
	messageBytes, byteCountReceived, err := Tcp.Receive(netConn, systemge.config.TcpTimeoutMs, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+connection_nodeName_topic+"\" message", err)
	}
	systemge.bytesReceived.Add(byteCountReceived)
	systemge.outgoingConnectionAttemptBytesReceived.Add(byteCountReceived)
	message, err := Message.Deserialize(messageBytes)
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
	messageBytes, byteCountReceived, err = Tcp.Receive(netConn, systemge.config.TcpTimeoutMs, systemge.config.IncomingMessageByteLimit)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to receive \""+connection_responsibleTopics_topic+"\" message", err)
	}
	systemge.bytesReceived.Add(byteCountReceived)
	systemge.outgoingConnectionAttemptBytesReceived.Add(byteCountReceived)
	message, err = Message.Deserialize(messageBytes)
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
	return systemge.newOutgoingConnection(netConn, endpointConfig, endpointName, topics)
}
