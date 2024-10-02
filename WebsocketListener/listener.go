package WebsocketListener

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketListener struct {
	name string

	isClosed    bool
	closedMutex sync.Mutex

	config        *Config.WebsocketListener
	ipRateLimiter *Tools.IpRateLimiter

	httpServer        *HTTPServer.HTTPServer
	connectionChannel chan *websocket.Conn

	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	eventHandler Event.Handler

	// metrics

	accepted atomic.Uint32
	failed   atomic.Uint32
	rejected atomic.Uint32
}

func New(name string, config *Config.WebsocketListener, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) (*WebsocketListener, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	listener := &WebsocketListener{
		name:         name,
		config:       config,
		blacklist:    blacklist,
		whitelist:    whitelist,
		eventHandler: eventHandler,
	}
	listener.httpServer = HTTPServer.New(listener.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: listener.config.TcpServerConfig,
		},
		whitelist, blacklist,
		map[string]http.HandlerFunc{
			listener.config.Pattern: server.getHTTPWebsocketUpgradeHandler(),
		},
		eventHandler,
	)
	if listener.config.IpRateLimiter != nil {
		listener.ipRateLimiter = Tools.NewIpRateLimiter(listener.config.IpRateLimiter)
	}
	if err := listener.httpServer.Start(); err != nil {
		return nil, err
	}

	return listener, nil
}
