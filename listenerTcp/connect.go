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

func EstablishConnection(config *configs.TcpBufferedReader, tcpClientConfig *configs.TcpClient, timeoutNs int64) (systemge.Connection[[]byte], error) {
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

type connector struct {
	tcpBufferedReaderConfig *configs.TcpBufferedReader
	tcpClientConfig         *configs.TcpClient
}

func (connector *connector) Connect(timeoutNs int64) (systemge.Connection[[]byte], error) {
	return EstablishConnection(connector.tcpBufferedReaderConfig, connector.tcpClientConfig, timeoutNs)
}

func (listener *TcpListener) GetConnector() systemge.Connector[[]byte, systemge.Connection[[]byte]] {
	connector := &connector{
		tcpBufferedReaderConfig: listener.tcpBufferedReaderConfig,
		tcpClientConfig: &configs.TcpClient{
			Port:   listener.config.Port,
			Domain: listener.config.Domain,
		},
	}
	if listener.config.TlsCertPath != "" {
		connector.tcpClientConfig.TlsCert = helpers.GetFileContent(listener.config.TlsCertPath)
	}
	return connector
}

func NewTcpClient(config *configs.TcpClient, timeoutNs int64) (net.Conn, error) {
	if config.TlsCert == "" {
		return net.Dial("tcp", config.Domain+":"+helpers.Uint16ToString(config.Port))
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
	return dialer.Dial("tcp", config.Domain+":"+helpers.Uint16ToString(config.Port))
}
