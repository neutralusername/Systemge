package Tcp

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"net"
)

func NewEndpoint(config *Config.TcpEndpoint) (net.Conn, error) {
	if config.TlsCert == "" {
		return net.Dial("tcp", config.Address)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(config.TlsCert)) {
		return nil, Error.New("Error adding certificate to root CAs", nil)
	}
	return tls.Dial("tcp", config.Address, &tls.Config{
		RootCAs:    rootCAs,
		ServerName: config.Domain,
	})
}
