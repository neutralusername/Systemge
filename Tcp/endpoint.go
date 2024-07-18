package Tcp

import (
	"Systemge/Error"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
)

type Endpoint struct {
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

func NewEndpoint(address, serverNameIndication, cert string) Endpoint {
	return Endpoint{
		address:              address,
		serverNameIndication: serverNameIndication,
		tlsCertificate:       cert,
	}
}

func (tcpEndpoint Endpoint) GetAddress() string {
	return tcpEndpoint.address
}

func (tcpEndpoint Endpoint) GetServerNameIndication() string {
	return tcpEndpoint.serverNameIndication
}

func (tcpEndpoint Endpoint) GetTlsCertificate() string {
	return tcpEndpoint.tlsCertificate
}

func (tcpEndpoint Endpoint) Dial() (net.Conn, error) {
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

func (tcpEndpoint Endpoint) Marshal() string {
	json, _ := json.Marshal(TcpEndpointData{
		Address:              tcpEndpoint.address,
		ServerNameIndication: tcpEndpoint.serverNameIndication,
		TlsCertificate:       tcpEndpoint.tlsCertificate,
	})
	return string(json)
}

func Unmarshal(data string) *Endpoint {
	var resolutionData TcpEndpointData
	json.Unmarshal([]byte(data), &resolutionData)
	endpoint := NewEndpoint(resolutionData.Address, resolutionData.ServerNameIndication, resolutionData.TlsCertificate)
	return &endpoint
}
