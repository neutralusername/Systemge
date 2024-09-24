package TcpSystemgeListener

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

type TcpSystemgeListener struct {
	name string

	isClosed    bool
	closedMutex sync.Mutex

	config        *Config.TcpSystemgeListener
	ipRateLimiter *Tools.IpRateLimiter

	tcpListener net.Listener
	acceptMutex sync.Mutex

	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	eventHandler func(*Event.Event) *Event.Event

	// metrics

	tcpSystemgeConnectionAttemptsTotal    atomic.Uint64
	tcpSystemgeConnectionAttemptsFailed   atomic.Uint64
	tcpSystemgeConnectionAttemptsRejected atomic.Uint64
	tcpSystemgeConnectionAttemptsAccepted atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeListener, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList) (*TcpSystemgeListener, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	server := &TcpSystemgeListener{
		name:      name,
		config:    config,
		blacklist: blacklist,
		whitelist: whitelist,
	}
	tcpListener, err := Tcp.NewListener(config.TcpServerConfig)
	if err != nil {
		return nil, err
	}
	server.tcpListener = tcpListener
	if config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiter)
	}
	return server, nil
}

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpSystemgeListener) Close() error {
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

func (listener *TcpSystemgeListener) GetStatus() int {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()
	if listener.isClosed {
		return Status.Stoped
	}
	return Status.Started
}

func (server *TcpSystemgeListener) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *TcpSystemgeListener) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *TcpSystemgeListener) GetListener() net.Listener {
	return server.tcpListener
}

func (server *TcpSystemgeListener) GetName() string {
	return server.name
}

func (server *TcpSystemgeListener) onEvent(event *Event.Event) *Event.Event {
	if server.eventHandler == nil {
		return event
	}
	return server.eventHandler(event)
}
func (server *TcpSystemgeListener) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.TcpSystemgeListener,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
}
