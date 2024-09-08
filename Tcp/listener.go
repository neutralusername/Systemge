package Tcp

import (
	"crypto/tls"
	"net"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
)

type Listener struct {
	config    *Config.TcpServer
	listener  net.Listener
	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList
}

func NewListener(config *Config.TcpServer, blacklist *Tools.AccessControlList, whitelist *Tools.AccessControlList) (*Listener, error) {
	if config.TlsCertPath == "" || config.TlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)))
		if err != nil {
			return nil, Error.New("Failed to listen on port: ", err)
		}
		server := &Listener{
			config:    config,
			listener:  listener,
			blacklist: blacklist,
			whitelist: whitelist,
		}
		return server, nil
	} else {
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
		server := &Listener{
			config:    config,
			listener:  listener,
			blacklist: blacklist,
			whitelist: whitelist,
		}
		return server, nil
	}
}

func (server *Listener) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *Listener) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *Listener) GetListener() net.Listener {
	return server.listener
}
