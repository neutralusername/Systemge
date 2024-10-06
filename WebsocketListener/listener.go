package WebsocketListener

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketListener struct {
	config *Config.WebsocketListener
	name   string

	instanceId string
	sessionId  string

	status      int
	statusMutex sync.Mutex
	stopChannel chan struct{}

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
		instanceId:        Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
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

func (listener *WebsocketListener) GetInstanceId() string {
	return listener.instanceId
}

func (listener *WebsocketListener) GetSessionId() string {
	return listener.sessionId
}

func (listener *WebsocketListener) GetStatus() int {
	return listener.status
}

func (server *WebsocketListener) GetName() string {
	return server.name
}

func (server *WebsocketListener) onEvent(listener *Event.Event) *Event.Event {
	listener.GetContext().Merge(server.GetContext())
	if server.eventHandler != nil {
		server.eventHandler(listener)
	}
	return listener
}
func (listener *WebsocketListener) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.TcpSystemgeListener,
		Event.ServiceName:       listener.name,
		Event.ServiceStatus:     Status.ToString(listener.GetStatus()),
		Event.ServiceInstanceId: listener.instanceId,
		Event.ServiceSessionId:  listener.sessionId,
		Event.Function:          Event.GetCallerFuncName(2),
	}
}
