package TcpSystemgeListener

import (
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpSystemgeListener struct {
	closed      bool
	closedMutex sync.Mutex

	config        *Config.TcpSystemgeListener
	ipRateLimiter *Tools.IpRateLimiter

	listener  net.Listener
	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	connectionId uint32

	// metrics

	connectionAttempts         atomic.Uint64
	failedConnectionAttempts   atomic.Uint64
	rejectedConnectionAttempts atomic.Uint64
	acceptedConnectionAttempts atomic.Uint64
}

func (server *TcpSystemgeListener) newListener(config *Config.TcpServer) (net.Listener, error) {
	if config.TlsCertPath == "" || config.TlsKeyPath == "" {
		listener, err := net.Listen("tcp", ":"+Helpers.IntToString(int(config.Port)))
		if err != nil {
			return nil, Error.New("Failed to listen on port: ", err)
		}
		return listener, nil
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
		return listener, nil
	}
}

func (server *TcpSystemgeListener) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *TcpSystemgeListener) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *TcpSystemgeListener) GetListener() net.Listener {
	return server.listener
}

func New(config *Config.TcpSystemgeListener, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList) (*TcpSystemgeListener, error) {
	if config == nil {
		return nil, Error.New("config is nil", nil)
	}
	if config.TcpServerConfig == nil {
		return nil, Error.New("listener is nil", nil)
	}
	server := &TcpSystemgeListener{
		config:    config,
		blacklist: blacklist,
		whitelist: whitelist,
	}
	tcpListener, err := server.newListener(config.TcpServerConfig)
	if err != nil {
		return nil, Error.New("failed to create listener", err)
	}
	server.listener = tcpListener
	if config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiter)
	}
	return server, nil
}

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpSystemgeListener) Close() error {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()
	if listener.closed {
		return Error.New("listener is already closed", nil)
	}
	listener.closed = true
	listener.listener.Close()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
	}
	return nil
}

func (listener *TcpSystemgeListener) GetStatus() int {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()
	if listener.closed {
		return Status.STOPPED
	}
	return Status.STARTED
}
