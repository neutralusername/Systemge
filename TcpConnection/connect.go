package TcpConnection

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
)

func EstablishConnection(config *Config.TcpSystemgeConnection, tcpClientConfig *Config.TcpClient) (*TcpConnection, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if tcpClientConfig == nil {
		return nil, errors.New("tcpClientConfig is nil")
	}

	netConn, err := NewTcpClient(tcpClientConfig)
	if err != nil {
		return nil, err
	}
	connection, err := New(config, netConn)
	if err != nil {
		netConn.Close()
		return nil, err
	}
	return connection, nil
}

func NewTcpClient(config *Config.TcpClient) (net.Conn, error) {
	if config.TlsCert == "" {
		return net.Dial("tcp", config.Address)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(config.TlsCert)) {
		return nil, errors.New("error adding certificate to root CAs")
	}
	return tls.Dial("tcp", config.Address, &tls.Config{
		RootCAs:    rootCAs,
		ServerName: config.Domain,
	})
}

/* _, err := Tcp.Write(netConn, Message.NewAsync(Message.TOPIC_NAME).Serialize(), config.TcpSendTimeoutMs)
if err != nil {
	return nil, err
}
messageReceiver := Tcp.NewBufferedMessageReader(netConn, config.IncomingMessageByteLimit, config.TcpReceiveTimeoutMs, config.TcpBufferBytes)
messageBytes, err := messageReceiver.ReadNextMessage()
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
	return nil, errors.New("expected \"" + Message.TOPIC_NAME + "\" message, but got \"" + message.GetTopic() + "\" message")
}
if maxServerNameLength > 0 && len(message.GetPayload()) > maxServerNameLength {
	return nil, errors.New("server name is too long")
}
if message.GetPayload() == "" {
	return nil, errors.New("server did not respond with a name")
} */
