package TcpServer

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
)

type TcpServer struct {
	port        uint16
	tlsCertPath string
	tlsKeyPath  string
}

func New(port uint16, tlsCertPath, tlsKeyPath string) TcpServer {
	return TcpServer{
		port:        port,
		tlsCertPath: tlsCertPath,
		tlsKeyPath:  tlsKeyPath,
	}
}

func (tlsEndpoint TcpServer) GetPort() uint16 {
	return tlsEndpoint.port
}

func (tlsEndpoint TcpServer) GetTlsCertPath() string {
	return tlsEndpoint.tlsCertPath
}

func (tlsEndpoint TcpServer) GetTlsKeyPath() string {
	return tlsEndpoint.tlsKeyPath
}

func (tlsEndpoint TcpServer) GetTlsListener() (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(tlsEndpoint.tlsCertPath, tlsEndpoint.tlsKeyPath)
	if err != nil {
		return nil, Error.New("Failed to load TLS certificate: ", err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, err := tls.Listen("tcp", ":"+Utilities.IntToString(int(tlsEndpoint.port)), config)
	if err != nil {
		return nil, Error.New("Failed to listen on port: ", err)
	}
	return listener, nil
}

func (tlsEndpoint TcpServer) GetTcpListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+Utilities.IntToString(int(tlsEndpoint.port)))
	if err != nil {
		return nil, Error.New("Failed to listen on port: ", err)
	}
	return listener, nil
}
