package Client

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/TCP"
	"crypto/tls"
	"crypto/x509"
)

func (client *Client) connectToBroker(resolution *Resolver.Resolution) (*serverConnection, error) {
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(resolution.Certificate)) {
		return nil, Error.New("Error adding certificate to root CAs", nil)
	}
	netConn, err := tls.Dial("tcp", resolution.Address, &tls.Config{
		RootCAs: rootCAs,
	})
	if err != nil {
		return nil, Error.New("Error connecting to message broker server", err)
	}
	err = TCP.Send(netConn, Message.NewAsync("connect", client.name, "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error sending connection request", err)
	}
	messageBytes, err := TCP.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Error.New("Error receiving connection response", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil {
		netConn.Close()
		return nil, Error.New("Invalid response \""+string(messageBytes)+"\"", nil)
	}
	if message.GetTopic() != "connected" {
		netConn.Close()
		return nil, Error.New("Invalid response topic \""+message.GetTopic()+"\"", nil)
	}
	return newServerConnection(netConn, resolution, client.logger), nil
}
