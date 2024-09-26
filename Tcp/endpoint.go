package Tcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"

	"github.com/neutralusername/Systemge/Config"
)

func NewClient(config *Config.TcpClient) (net.Conn, error) {
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
