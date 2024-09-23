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

	closed      bool
	closedMutex sync.Mutex

	config        *Config.TcpSystemgeListener
	ipRateLimiter *Tools.IpRateLimiter

	listener  net.Listener
	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	connectionId uint32

	onErrorHandler   func(*Event.Event) *Event.Event
	onWarningHandler func(*Event.Event) *Event.Event
	onInfoHandler    func(*Event.Event) *Event.Event

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
		return errors.New("tcpSystemgeListener is already closed")
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

func (server *TcpSystemgeListener) onError(event *Event.Event) *Event.Event {
	if server.onErrorHandler != nil {
		return server.onErrorHandler(event)
	}
	return event
}

func (server *TcpSystemgeListener) onWarning(event *Event.Event) *Event.Event {
	if server.onWarningHandler != nil {
		return server.onWarningHandler(event)
	}
	return event
}

func (server *TcpSystemgeListener) onInfo(event *Event.Event) *Event.Event {
	if server.onInfoHandler != nil {
		return server.onInfoHandler(event)
	}
	return event
}

func (server *TcpSystemgeListener) GetServerContext() Event.Context {
	ctx := Event.Context{
		Event.Service:       Event.TcpSystemgeListener,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
	return ctx
}

func (server *TcpSystemgeListener) GetName() string {
	return server.name
}
