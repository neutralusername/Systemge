package Tcp

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Helpers"
	"Systemge/Tools"
	"crypto/tls"
	"net"
)

type Server struct {
	config    *Config.TcpServer
	listener  net.Listener
	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList
}

func NewServer(config *Config.TcpServer) (*Server, error) {
	if config.TlsCertPath == "" || config.TlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)))
		if err != nil {
			return nil, Error.New("Failed to listen on port: ", err)
		}
		server := &Server{
			config:    config,
			listener:  listener,
			blacklist: Tools.NewAccessControlList(config.Blacklist),
			whitelist: Tools.NewAccessControlList(config.Whitelist),
		}
		return server, nil
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
	server := &Server{
		config:    config,
		listener:  listener,
		blacklist: Tools.NewAccessControlList(config.Blacklist),
		whitelist: Tools.NewAccessControlList(config.Whitelist),
	}
	return server, nil
}

func (server *Server) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *Server) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *Server) GetListener() net.Listener {
	return server.listener
}
