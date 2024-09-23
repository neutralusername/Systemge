package TcpSystemgeListener

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
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
		return nil, Event.New("config is nil", nil)
	}
	if config.TcpServerConfig == nil {
		return nil, Event.New("listener is nil", nil)
	}
	server := &TcpSystemgeListener{
		config:    config,
		blacklist: blacklist,
		whitelist: whitelist,
	}
	tcpListener, err := Tcp.NewListener(config.TcpServerConfig)
	if err != nil {
		return nil, Event.New("failed to create listener", err)
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
		return Event.New("listener is already closed", nil)
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
		return Status.Stoped
	}
	return Status.Started
}
