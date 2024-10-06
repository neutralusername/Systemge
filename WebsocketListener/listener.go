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
)

type WebsocketListener struct {
	name string

	status      int
	statusMutex sync.Mutex
	stopChannel chan bool

	config *Config.WebsocketListener

	httpServer        *HTTPServer.HTTPServer
	connectionChannel chan *websocket.Conn

	eventHandler Event.Handler

	// metrics

	clientsAccepted atomic.Uint64
	clientsFailed   atomic.Uint64
	clientsRejected atomic.Uint64
}

func New(name string, config *Config.WebsocketListener, eventHandler Event.Handler) (*WebsocketListener, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	listener := &WebsocketListener{
		name:              name,
		config:            config,
		eventHandler:      eventHandler,
		status:            Status.Stopped,
		connectionChannel: make(chan *websocket.Conn),
	}
	listener.httpServer = HTTPServer.New(listener.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: listener.config.TcpServerConfig,
		},
		nil, nil,
		map[string]http.HandlerFunc{
			listener.config.Pattern: listener.getHTTPWebsocketUpgradeHandler(),
		},
		eventHandler,
	)

	return listener, nil
}

func (listener *WebsocketListener) GetStatus() int {
	return listener.status
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
