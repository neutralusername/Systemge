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

func (tlsServer TcpServer) GetPort() uint16 {
	return tlsServer.port
}

func (tlsServer TcpServer) GetTlsCertPath() string {
	return tlsServer.tlsCertPath
}

func (tlsServer TcpServer) GetTlsKeyPath() string {
	return tlsServer.tlsKeyPath
}

func (tlsServer TcpServer) GetListener() (net.Listener, error) {
	if tlsServer.tlsCertPath == "" || tlsServer.tlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Utilities.IntToString(int(tlsServer.port)))
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
	listener, err := tls.Listen("tcp", ":"+Utilities.IntToString(int(tlsServer.port)), config)
	if err != nil {
		return nil, Error.New("Failed to listen on port: ", err)
	}
	return listener, nil
}
