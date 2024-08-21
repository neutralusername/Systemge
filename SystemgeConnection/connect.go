package SystemgeConnection

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/Tcp"
)

func EstablishConnection(config *Config.SystemgeConnection, endpointConfig *Config.TcpEndpoint, clientName string, maxServerNameLength int, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) (*SystemgeConnection, error) {
	if config == nil {
		return nil, Error.New("Config is nil", nil)
	}
	netConn, err := Tcp.NewClient(endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to "+endpointConfig.Address, err)
	}
	connection, err := clientHandshake(config, clientName, maxServerNameLength, netConn, messageHandler)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Failed to handshake with "+endpointConfig.Address, err)
	}
	return connection, nil
}

func clientHandshake(config *Config.SystemgeConnection, clientName string, maxServerNameLength int, netConn net.Conn, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) (*SystemgeConnection, error) {
	name := ""
	connection := New(config, netConn, "", messageHandler)
	err := connection.AsyncMessage(Message.TOPIC_NAME, clientName)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	message, err := connection.GetNextMessage()
	if err != nil {
		return nil, Error.New("Failed to process \""+Message.TOPIC_NAME+"\" message", err)
	}
	SystemgeMessageHandler.New(SystemgeMessageHandler.AsyncMessageHandlers{
		Message.TOPIC_NAME: func(message *Message.Message) {
			if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
				return
			}
			name = message.GetPayload()
		},
	}, nil).HandleAsyncMessage(message)
	if name == "" {
		return nil, Error.New("Server did not respond with a name", nil)
	}
	connection.name = name
	return connection, nil
}
