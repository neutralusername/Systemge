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
	channel := make(chan struct{})
	conn := New(config, netConn, "", SystemgeMessageHandler.New(SystemgeMessageHandler.AsyncMessageHandlers{
		Message.TOPIC_NAME: func(message *Message.Message) {
			if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
				return
			}
			name = message.GetPayload()
			close(channel)
		},
	}, nil))
	err := conn.AsyncMessage(Message.TOPIC_NAME, clientName)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	<-channel
	if name == "" {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	connection := New(config, netConn, name, messageHandler)
	connection.tcpBuffer = conn.tcpBuffer
	return connection, nil
}
