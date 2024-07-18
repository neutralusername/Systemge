package Tcp

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Helpers"
	"crypto/tls"
	"net"
)

func NewServer(config Config.TcpServer) (net.Listener, error) {
	if config.TlsCertPath == "" || config.TlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)))
		if err != nil {
			return nil, Error.New("Failed to listen on port: ", err)
		}
		return listener, nil
	}
	cert, err := tls.LoadX509KeyPair(config.TlsCertPath, config.TlsKeyPath)
	if err != nil {
		return nil, Error.New("Failed to load TLS certificate: ", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, err := tls.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)), tlsConfig)
	if err != nil {
		return nil, Error.New("Failed to listen on port: ", err)
	}
	return listener, nil
}
