package TcpSystemgeConnection

import (
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
)

func EstablishConnection(config *Config.TcpSystemgeConnection, tcpClientConfig *Config.TcpClient, clientName string, maxServerNameLength int) (SystemgeConnection.SystemgeConnection, error) {
	if config == nil {
		return nil, Event.New("Config is nil", nil)
	}
	netConn, err := Tcp.NewClient(tcpClientConfig)
	if err != nil {
		return nil, Event.New("Failed to establish connection to "+tcpClientConfig.Address, err)
	}
	connection, err := clientHandshake(config, clientName, maxServerNameLength, netConn)
	if err != nil {
		netConn.Close()
		return nil, Event.New("Failed to handshake with "+tcpClientConfig.Address, err)
	}
	return connection, nil
}

func clientHandshake(config *Config.TcpSystemgeConnection, clientName string, maxServerNameLength int, netConn net.Conn) (*TcpSystemgeConnection, error) {
	_, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, clientName).Serialize(), config.TcpSendTimeoutMs)
	if err != nil {
		return nil, Event.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	messageReceiver := Tcp.NewBufferedMessageReceiver(netConn, config.IncomingMessageByteLimit, config.TcpReceiveTimeoutMs, config.TcpBufferBytes)
	messageBytes, _, err := messageReceiver.ReceiveNextMessage()
	if err != nil {
		return nil, Event.New("Failed to receive response", err)
	}
	if len(messageBytes) == 0 {
		return nil, Event.New("Received empty message", nil)
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
		return nil, Event.New("Failed to deserialize response \""+string(filteresMessageBytes)+"\"", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Event.New("Expected \""+Message.TOPIC_NAME+"\" message, but got \""+message.GetTopic()+"\" message", nil)
	}
	if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
		return nil, Event.New("Server name is too long", nil)
	}
	if message.GetPayload() == "" {
		return nil, Event.New("Server did not respond with a name", nil)
	}
	return New(message.GetPayload(), config, netConn, messageReceiver), nil
}
