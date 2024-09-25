package TcpSystemgeConnect

import (
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

func EstablishConnection(config *Config.TcpSystemgeConnection, tcpClientConfig *Config.TcpClient, clientName string, maxServerNameLength int, eventHandler Event.Handler) (SystemgeConnection.SystemgeConnection, error) {
	if config == nil {
		return nil, errors.New("Config is nil")
	}
	netConn, err := Tcp.NewClient(tcpClientConfig)
	if err != nil {
		return nil, err
	}
	connection, err := clientHandshake(config, clientName, maxServerNameLength, netConn, eventHandler)
	if err != nil {
		netConn.Close()
		return nil, err
	}
	return connection, nil
}

func clientHandshake(config *Config.TcpSystemgeConnection, clientName string, maxServerNameLength int, netConn net.Conn, eventHandler Event.Handler) (*TcpSystemgeConnection.TcpSystemgeConnection, error) {
	_, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, clientName).Serialize(), config.TcpSendTimeoutMs)
	if err != nil {
		return nil, err
	}
	messageReceiver := Tcp.NewBufferedMessageReceiver(netConn, config.IncomingMessageByteLimit, config.TcpReceiveTimeoutMs, config.TcpBufferBytes)
	messageBytes, _, err := messageReceiver.ReceiveNextMessage()
	if err != nil {
		return nil, err
	}
	if len(messageBytes) == 0 {
		return nil, errors.New("received empty message")
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
		return nil, err
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, errors.New("Expected \"" + Message.TOPIC_NAME + "\" message, but got \"" + message.GetTopic() + "\" message")
	}
	if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
		return nil, errors.New("Server name is too long")
	}
	if message.GetPayload() == "" {
		return nil, errors.New("Server did not respond with a name")
	}
	return TcpSystemgeConnection.New(message.GetPayload(), config, netConn, messageReceiver, eventHandler), nil
}
