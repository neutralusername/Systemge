package Tcp

import (
	"crypto/tls"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Helpers"
)

func NewListener(config *Config.TcpServer) (net.Listener, error) {
	if config.TlsCertPath == "" || config.TlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)))
		if err != nil {
			return nil, err
		}
		return listener, nil
	} else {
		cert, err := tls.LoadX509KeyPair(config.TlsCertPath, config.TlsKeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		listener, err := tls.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)), tlsConfig)
		if err != nil {
			return nil, err
		}
		return listener, nil
	}
}
