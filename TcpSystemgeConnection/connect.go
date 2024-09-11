package TcpSystemgeConnection

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
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
	_, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, clientName).Serialize(), config.TcpSendTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	messageBytes, _, err := Tcp.Receive(netConn, config.TcpReceiveTimeoutMs, 4096)
	if err != nil {
		return nil, Error.New("Failed to receive response", err)
	}
	if len(messageBytes) == 0 {
		return nil, Error.New("Received empty message", nil)
	}
	filteresMessageBytes := []byte{}
	for _, b := range messageBytes {
		if b == Tcp.HEARTBEAT {
			continue
		}
		if b == Tcp.ENDOFMESSAGE {
			continue
		}
		filteresMessageBytes = append(filteresMessageBytes, b)
	}
	message, err := Message.Deserialize(filteresMessageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize response", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Expected \""+Message.TOPIC_NAME+"\" message, but got \""+message.GetTopic()+"\" message", nil)
	}
	if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
		return nil, Error.New("Server name is too long", nil)
	}
	if message.GetPayload() == "" {
		return nil, Error.New("Server did not respond with a name", nil)
	}
	return New(message.GetPayload(), config, netConn), nil
}
