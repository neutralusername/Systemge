package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"crypto/tls"
	"crypto/x509"
	"net"
)

func (client *Client) tcpDial(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

func (client *Client) tlsDial(address string, tlsCertificate string) (net.Conn, error) {
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(tlsCertificate)) {
		return nil, Utilities.NewError("Error adding certificate to root CAs", nil)
	}
	return tls.Dial("tcp", address, &tls.Config{
		RootCAs: rootCAs,
	})
}

func (client *Client) tcpExchange(netConn net.Conn, message *Message.Message) (*Message.Message, error) {
	err := Utilities.TcpSend(netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error sending message", err)
	}
	responseBytes, err := Utilities.TcpReceive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error receiving response", err)
	}
	message = Message.Deserialize(responseBytes)
	if message == nil {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response \""+string(responseBytes)+"\"", nil)
	}
	return message, nil
}
