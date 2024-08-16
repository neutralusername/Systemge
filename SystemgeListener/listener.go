package SystemgeListener

import (
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener struct {
	config        *Config.SystemgeListener
	tcpListener   *Tcp.Listener
	ipRateLimiter *Tools.IpRateLimiter

	connectionId uint32

	// metrics

	connectionAttempts  atomic.Uint32
	failedConnections   atomic.Uint32
	rejectedConnections atomic.Uint32
	acceptedConnections atomic.Uint32
}

func New(config *Config.SystemgeListener) *SystemgeListener {
	if config == nil {
		panic("config is nil")
	}
	tcpListener, err := Tcp.NewListener(config.ListenerConfig)
	if err != nil {
		panic(err)
	}
	listener := &SystemgeListener{
		config:      config,
		tcpListener: tcpListener,
	}
	if config.IpRateLimiter != nil {
		listener.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiter)
	}
	return listener
}

func (listener *SystemgeListener) Close() {
	listener.tcpListener.GetListener().Close()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Stop()
	}
}
