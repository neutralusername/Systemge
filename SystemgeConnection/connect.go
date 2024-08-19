package SystemgeConnection

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func EstablishConnection(config *Config.SystemgeConnection, endpointConfig *Config.TcpEndpoint, clientName string, maxServerNameLength int) (*SystemgeConnection, error) {
	if config == nil {
		return nil, Error.New("Config is nil", nil)
	}
	netConn, err := Tcp.NewClient(endpointConfig)
	if err != nil {
		return nil, Error.New("Failed to establish connection to "+endpointConfig.Address, err)
	}
	connection, err := clientHandshake(config, clientName, maxServerNameLength, netConn)
	if err != nil {
		return nil, Error.New("Failed to handshake with "+endpointConfig.Address, err)
	}
	return connection, nil
}

func clientHandshake(config *Config.SystemgeConnection, clientName string, maxServerNameLength int, netConn net.Conn) (*SystemgeConnection, error) {
	_, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, clientName).Serialize(), config.TcpSendTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	messageBytes, _, err := Tcp.Receive(netConn, config.TcpReceiveTimeoutMs, config.TcpBufferBytes)
	if err != nil {
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	messageBytes = messageBytes[:len(messageBytes)-1]
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
		return nil, Error.New("Received server name \""+message.GetPayload()+"\" exceeds maximum size of "+Helpers.IntToString(maxServerNameLength), nil)
	}
	if len(message.GetPayload()) == 0 {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	return New(config, netConn, message.GetPayload()), nil
}
