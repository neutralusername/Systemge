package TcpEndpoint

import (
	"Systemge/Error"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
)

type TcpEndpoint struct {
	address              string
	serverNameIndication string
	tlsCertificate       string
}

type TcpEndpointData struct {
	Name                 string `json:"name"`
	Address              string `json:"port"`
	ServerNameIndication string `json:"serverNameIndication"`
	TlsCertificate       string `json:"tlsCertificate"`
}

func New(address, serverNameIndication, cert string) TcpEndpoint {
	return TcpEndpoint{
		address:              address,
		serverNameIndication: serverNameIndication,
		tlsCertificate:       cert,
	}
}

func (tcpEndpoint TcpEndpoint) GetAddress() string {
	return tcpEndpoint.address
}

func (tcpEndpoint TcpEndpoint) GetServerNameIndication() string {
	return tcpEndpoint.serverNameIndication
}

func (tcpEndpoint TcpEndpoint) GetTlsCertificate() string {
	return tcpEndpoint.tlsCertificate
}

func (tcpEndpoint TcpEndpoint) Dial() (net.Conn, error) {
	if tcpEndpoint.tlsCertificate == "" {
		return net.Dial("tcp", tcpEndpoint.address)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(tcpEndpoint.tlsCertificate)) {
		return nil, Error.New("Error adding certificate to root CAs", nil)
	}
	return tls.Dial("tcp", tcpEndpoint.address, &tls.Config{
		RootCAs:    rootCAs,
		ServerName: tcpEndpoint.serverNameIndication,
	})
}

func (tcpEndpoint TcpEndpoint) Marshal() string {
	json, _ := json.Marshal(TcpEndpointData{
		Address:              tcpEndpoint.address,
		ServerNameIndication: tcpEndpoint.serverNameIndication,
		TlsCertificate:       tcpEndpoint.tlsCertificate,
	})
	return string(json)
}

func Unmarshal(data string) *TcpEndpoint {
	var resolutionData TcpEndpointData
	json.Unmarshal([]byte(data), &resolutionData)
	resolution := New(resolutionData.Address, resolutionData.ServerNameIndication, resolutionData.TlsCertificate)
	return &resolution
}
