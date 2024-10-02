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
	"github.com/neutralusername/Systemge/Status"
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
			listener.config.Pattern: listener.getHTTPWebsocketUpgradeHandler(),
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

func (listener *WebsocketListener) Close() error {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()

	if event := listener.onEvent(Event.NewInfo(
		Event.ServiceStopping,
		"stopping tcpSystemgeListener",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStopping,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if listener.isClosed {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"tcpSystemgeListener is already closed",
			Event.Context{
				Event.Circumstance: Event.ServiceStopping,
			},
		))
		return errors.New("tcpSystemgeListener is already closed")
	}

	listener.isClosed = true
	listener.httpServer.Stop()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
	}

	listener.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"tcpSystemgeListener stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStopping,
		},
	))

	return nil
}

func (listener *WebsocketListener) GetStatus() int {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()
	if listener.isClosed {
		return Status.Stopped
	}
	return Status.Started
}

func (server *WebsocketListener) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *WebsocketListener) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *WebsocketListener) GetName() string {
	return server.name
}

func (server *WebsocketListener) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *WebsocketListener) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.TcpSystemgeListener,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
}
