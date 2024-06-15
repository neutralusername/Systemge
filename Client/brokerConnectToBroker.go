package Client

import (
	"Systemge/Message"
	"Systemge/Resolver"
	"Systemge/Utilities"
	"crypto/tls"
	"crypto/x509"
)

func (client *Client) connectToBroker(resolution *Resolver.Resolution) (*brokerConnection, error) {
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(resolution.Certificate)) {
		return nil, Utilities.NewError("Error adding certificate to root CAs", nil)
	}
	netConn, err := tls.Dial("tcp", resolution.Address, &tls.Config{
		RootCAs: rootCAs,
	})
	if err != nil {
		return nil, Utilities.NewError("Error connecting to message broker server", err)
	}
	err = Utilities.Send(netConn, Message.NewAsync("connect", client.name, "").Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error sending connection request", err)
	}
	messageBytes, err := Utilities.Receive(netConn, DEFAULT_TCP_TIMEOUT)
	if err != nil {
		netConn.Close()
		return nil, Utilities.NewError("Error receiving connection response", err)
	}
	message := Message.Deserialize(messageBytes)
	if message == nil {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response \""+string(messageBytes)+"\"", nil)
	}
	if message.GetTopic() != "connected" {
		netConn.Close()
		return nil, Utilities.NewError("Invalid response topic \""+message.GetTopic()+"\"", nil)
	}
	return newBrokerConnection(netConn, resolution, client.logger), nil
}
