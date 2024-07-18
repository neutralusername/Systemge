package Tcp

import (
	"Systemge/Error"
	"Systemge/Helpers"
	"crypto/tls"
	"net"
)

type Server struct {
	port        uint16
	tlsCertPath string
	tlsKeyPath  string
}

func NewServer(port uint16, tlsCertPath, tlsKeyPath string) Server {
	return Server{
		port:        port,
		tlsCertPath: tlsCertPath,
		tlsKeyPath:  tlsKeyPath,
	}
}

func (tlsServer Server) GetPort() uint16 {
	return tlsServer.port
}

func (tlsServer Server) GetTlsCertPath() string {
	return tlsServer.tlsCertPath
}

func (tlsServer Server) GetTlsKeyPath() string {
	return tlsServer.tlsKeyPath
}

func (tlsServer Server) GetListener() (net.Listener, error) {
	if tlsServer.tlsCertPath == "" || tlsServer.tlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Helpers.IntToString(int(tlsServer.port)))
		if err != nil {
			return nil, Error.New("Failed to listen on port: ", err)
		}
		return listener, nil
	}
	cert, err := tls.LoadX509KeyPair(tlsServer.tlsCertPath, tlsServer.tlsKeyPath)
	if err != nil {
		return nil, Error.New("Failed to load TLS certificate: ", err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, err := tls.Listen("tcp", ":"+Helpers.IntToString(int(tlsServer.port)), config)
	if err != nil {
		return nil, Error.New("Failed to listen on port: ", err)
	}
	return listener, nil
}
