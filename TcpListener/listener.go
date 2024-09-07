package TcpListener

import (
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpListener struct {
	config        *Config.TcpSystemgeListener
	tcpListener   *Tcp.Listener
	ipRateLimiter *Tools.IpRateLimiter

	connectionId uint32

	// metrics

	connectionAttempts  atomic.Uint64
	failedConnections   atomic.Uint64
	rejectedConnections atomic.Uint64
	acceptedConnections atomic.Uint64
}

func New(config *Config.TcpSystemgeListener) (*TcpListener, error) {
	if config == nil {
		return nil, Error.New("config is nil", nil)
	}
	if config.TcpServerConfig == nil {
		return nil, Error.New("listener is nil", nil)
	}
	tcpListener, err := Tcp.NewListener(config.TcpServerConfig)
	if err != nil {
		return nil, Error.New("failed to create listener", err)
	}
	listener := &TcpListener{
		config:      config,
		tcpListener: tcpListener,
	}
	if config.IpRateLimiter != nil {
		listener.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiter)
	}
	return listener, nil
}

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpListener) Close() {
	listener.tcpListener.GetListener().Close()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
	}
}

func (listener *TcpListener) GetBlacklist() *Tools.AccessControlList {
	return listener.tcpListener.GetBlacklist()
}

func (listener *TcpListener) GetWhitelist() *Tools.AccessControlList {
	return listener.tcpListener.GetWhitelist()
}
