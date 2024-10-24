package listenerTcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"time"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionTcp"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
)

func Connect(
	config *configs.TcpBufferedReader,
	tcpClientConfig *configs.TcpClient,
	timeoutNs int64,
) (systemge.Connection[[]byte], error) {

	if config == nil {
		return nil, errors.New("config is nil")
	}
	if tcpClientConfig == nil {
		return nil, errors.New("tcpClientConfig is nil")
	}

	netConn, err := NewTcpClient(tcpClientConfig, timeoutNs)
	if err != nil {
		return nil, err
	}
	connection, err := connectionTcp.New(config, netConn)
	if err != nil {
		netConn.Close()
		return nil, err
	}
	return connection, nil
}

func NewTcpClient(config *configs.TcpClient, timeoutNs int64) (net.Conn, error) {
	if config.Ip != "" {
		ip, err := net.LookupIP(config.Domain)
		if err != nil {
			return nil, err
		}
		config.Ip = ip[0].String()
	}

	if config.TlsCert == "" {
		return net.Dial("tcp", config.Ip+":"+helpers.Uint16ToString(config.Port))
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(config.TlsCert)) {
		return nil, errors.New("error adding certificate to root CAs")
	}
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: time.Duration(timeoutNs) * time.Nanosecond,
		},
		Config: &tls.Config{
			RootCAs:    rootCAs,
			ServerName: config.Domain,
		},
	}
	return dialer.Dial("tcp", config.Ip+":"+helpers.Uint16ToString(config.Port))
}
