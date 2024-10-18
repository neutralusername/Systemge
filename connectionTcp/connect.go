package connectionTcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"time"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
)

func EstablishConnection(config *configs.TcpBufferedReader, tcpClientConfig *configs.TcpClient, lifetimeNs int64) (systemge.Connection[[]byte], error) {
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
	connection, err := New(config, netConn, lifetimeNs)
	if err != nil {
		netConn.Close()
		return nil, err
	}
	return connection, nil
}

func NewTcpClient(config *configs.TcpClient) (net.Conn, error) {
	if config.TlsCert == "" {
		return net.Dial("tcp", config.Address)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(config.TlsCert)) {
		return nil, errors.New("error adding certificate to root CAs")
	}
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: time.Duration(config.DialTimeoutNs) * time.Nanosecond,
		},
		Config: &tls.Config{
			RootCAs:    rootCAs,
			ServerName: config.Domain,
		},
	}
	return dialer.Dial("tcp", config.Address)
}
