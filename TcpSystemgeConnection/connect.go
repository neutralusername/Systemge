package TcpSystemgeConnection

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func EstablishConnection(config *Config.TcpSystemgeConnection, endpointConfig *Config.TcpClient, clientName string, maxServerNameLength int) (SystemgeConnection.SystemgeConnection, error) {
	if config == nil {
		return nil, Error.New("Config is nil", nil)
	}
	netConn, err := Tcp.NewClient(endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to "+endpointConfig.Address, err)
	}
	connection, err := clientHandshake(config, clientName, maxServerNameLength, netConn)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to handshake with "+endpointConfig.Address, err)
	}
	return connection, nil
}

func clientHandshake(config *Config.TcpSystemgeConnection, clientName string, maxServerNameLength int, netConn net.Conn) (*TcpConnection, error) {
	connection := New(clientName+"_connectionAttempt", config, netConn)
	err := connection.AsyncMessage(Message.TOPIC_NAME, clientName)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	message, err := connection.GetNextMessage()
	if err != nil {
		return nil, Error.New("Failed to get next message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Expected \""+Message.TOPIC_NAME+"\" message, but got \""+message.GetTopic()+"\" message", nil)
	}
	name := ""
	SystemgeConnection.NewConcurrentMessageHandler(SystemgeConnection.AsyncMessageHandlers{
		Message.TOPIC_NAME: func(connection SystemgeConnection.SystemgeConnection, message *Message.Message) {
			if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
				return
			}
			name = message.GetPayload()
		},
	}, nil, nil, nil).HandleAsyncMessage(connection, message)
	if name == "" {
		return nil, Error.New("Server did not respond with a name", nil)
	}
	connection.name = name
	if config.InfoLoggerPath != "" {
		connection.infoLogger = Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		connection.warningLogger = Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		connection.errorLogger = Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	return connection, nil
}
