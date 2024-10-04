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

	status      int
	statusMutex sync.Mutex

	config        *Config.WebsocketListener
	ipRateLimiter *Tools.IpRateLimiter

	httpServer        *HTTPServer.HTTPServer
	connectionChannel chan *websocket.Conn

	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList

	eventHandler Event.Handler

	// metrics

	clientsAccepted atomic.Uint64
	clientsFailed   atomic.Uint64
	clientsRejected atomic.Uint64
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

	return listener, nil
}

func (listener *WebsocketListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if event := listener.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting tcpSystemgeListener",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStarting,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if listener.status == Status.Started {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"tcpSystemgeListener is already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStarting,
			},
		))
		return errors.New("tcpSystemgeListener is already started")
	}

	listener.status = Status.Pending
	if listener.config.IpRateLimiter != nil {
		listener.ipRateLimiter = Tools.NewIpRateLimiter(listener.config.IpRateLimiter)
	}
	if err := listener.httpServer.Start(); err != nil {
		listener.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServiceStarting,
			},
		))
		listener.ipRateLimiter.Close()
		listener.ipRateLimiter = nil
		listener.status = Status.Stopped
		return err
	}

	listener.status = Status.Started
	listener.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"tcpSystemgeListener started",
		Event.Context{
			Event.Circumstance: Event.ServiceStarting,
		},
	))
	return nil
}

func (listener *WebsocketListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if event := listener.onEvent(Event.NewInfo(
		Event.ServiceStoping,
		"stopping websocketListener",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStoping,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if listener.status == Status.Stopped {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"websocketListener is already stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStoping,
			},
		))
		return errors.New("websocketListener is already stopped")
	}

	listener.status = Status.Pending
	listener.httpServer.Stop()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
		listener.ipRateLimiter = nil
	}

	listener.onEvent(Event.NewInfoNoOption(
		Event.ServiceStoped,
		"websocketListener stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStoping,
		},
	))

	listener.status = Status.Stopped
	return nil
}

func (listener *WebsocketListener) GetStatus() int {
	return listener.status
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
