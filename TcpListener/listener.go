package TcpListener

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpListener struct {
	closed      bool
	closedMutex sync.Mutex

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
func (listener *TcpListener) Close() error {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()
	if listener.closed {
		return Error.New("listener is already closed", nil)
	}
	listener.closed = true
	listener.tcpListener.GetListener().Close()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
	}
	return nil
}

func (listener *TcpListener) GetStatus() int {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()
	if listener.closed {
		return Status.STOPPED
	}
	return Status.STARTED
}

func (listener *TcpListener) GetBlacklist() *Tools.AccessControlList {
	return listener.tcpListener.GetBlacklist()
}

func (listener *TcpListener) GetWhitelist() *Tools.AccessControlList {
	return listener.tcpListener.GetWhitelist()
}

func (listener *TcpListener) GetDefaultCommands() Commands.Handlers {
	blacklistCommands := listener.tcpListener.GetBlacklist().GetCommands()
	whitelistCommands := listener.tcpListener.GetWhitelist().GetCommands()
	commands := Commands.Handlers{}
	for key, value := range blacklistCommands {
		commands["blacklist_"+key] = value
	}
	for key, value := range whitelistCommands {
		commands["whitelist_"+key] = value
	}
	commands["close"] = func(args []string) (string, error) {
		listener.Close()
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(listener.GetStatus()), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := listener.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["retrieveMetrics"] = func(args []string) (string, error) {
		metrics := listener.RetrieveMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	return commands
}
