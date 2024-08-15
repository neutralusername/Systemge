package SystemgeClient

import (
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
func (client *SystemgeClient) attemptServerConnection(endpointConfig *Config.TcpEndpoint) (*serverConnection, error) {
	address, err := Helpers.NormalizeAddress(endpointConfig.Address)
	if err != nil {
		return nil, Error.New("Failed to normalize address \""+endpointConfig.Address+"\"", err)
	}
	endpointConfig.Address = address

	serverConnectionAttempt := &serverConnectionAttempt{
		endpointConfig: endpointConfig,
	}

	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting connection attempts to endpoint \""+endpointConfig.Address+"\"", nil).Error())
	}
	for {
		select {
		case <-client.stopChannel:
			return nil, Error.New("server connection attempt to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\" was stopped because systemge was stopped", nil)
		default:
			if maxConnectionAttempts := client.config.MaxConnectionAttempts; maxConnectionAttempts > 0 && serverConnectionAttempt.attempts >= maxConnectionAttempts {
				return nil, Error.New("Max attempts reached to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", nil)
			}
			if serverConnectionAttempt.isAborted {
				return nil, Error.New("server connection attempt to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\" was aborted", nil)
			}
			if serverConnectionAttempt.attempts > 0 {
				time.Sleep(time.Duration(client.config.ConnectionAttemptDelayMs) * time.Millisecond)
			}
			if infoLogger := client.infoLogger; infoLogger != nil {
				infoLogger.Log(Error.New("Attempt #"+Helpers.Uint64ToString(serverConnectionAttempt.attempts)+" to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", nil).Error())
			}
			if serverConnnection, err := client.handleServerConnectionHandshake(serverConnectionAttempt); err != nil {
				client.connectionAttemptsFailed.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed attempt #"+Helpers.Uint64ToString(serverConnectionAttempt.attempts)+" to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\"", err).Error())
				}
				serverConnectionAttempt.attempts++
				continue
			} else {
				client.connectionAttemptsSuccessful.Add(1)
				if infoLogger := client.infoLogger; infoLogger != nil {
					infoLogger.Log(Error.New("Succeded attempt #"+Helpers.Uint64ToString(serverConnectionAttempt.attempts)+" to connect to endpoint \""+serverConnectionAttempt.endpointConfig.Address+"\" with name \""+serverConnnection.name+"\"", nil).Error())
				}
				return serverConnnection, nil
			}
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
	tcpBuffer := serverConnection.tcpBuffer
	serverConnection = client.newServerConnection(netConn, serverConnectionAttempt.endpointConfig, endpointName)
	serverConnection.tcpBuffer = tcpBuffer
	return serverConnection, nil
}
