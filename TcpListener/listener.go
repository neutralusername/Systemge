package TcpListener

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpListener struct {
	name string

	isClosed    bool
	closedMutex sync.Mutex

	config        *Config.TcpSystemgeListener
	ipRateLimiter *Tools.IpRateLimiter

	tcpListener net.Listener
	acceptMutex sync.Mutex

	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	eventHandler Event.Handler

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeListener, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, ipRateLimiter *Tools.IpRateLimiter) (*TcpListener, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	server := &TcpListener{
		name:          name,
		config:        config,
		blacklist:     blacklist,
		whitelist:     whitelist,
		ipRateLimiter: ipRateLimiter,
	}
	tcpListener, err := Tcp.NewListener(config.TcpServerConfig)
	if err != nil {
		return nil, err
	}
	server.tcpListener = tcpListener
	return server, nil
}

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpListener) Close() error {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()

	if listener.isClosed {
		return errors.New("tcpSystemgeListener is already closed")
	}

	listener.isClosed = true
	listener.tcpListener.Close()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
	}

	return nil
}

func (listener *TcpListener) GetStatus() int {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()

	if listener.isClosed {
		return Status.Stopped
	}
	return Status.Started
}

func (server *TcpListener) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *TcpListener) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *TcpListener) GetIpRateLimiter() *Tools.IpRateLimiter {
	return server.ipRateLimiter
}

func (server *TcpListener) GetName() string {
	return server.name
}
