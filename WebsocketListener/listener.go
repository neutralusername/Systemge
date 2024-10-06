package WebsocketListener

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
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

	httpServer *HTTPServer.HTTPServer

	anySemaphore *Tools.AnySemaphore

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
	ClientsRejected atomic.Uint64
}

type acceptRequest struct {
	upgraderResponseChannel chan *upgraderResponse
	timeoutMs               uint32
	triggered               sync.WaitGroup
	timedOut                bool
}

type upgraderResponse struct {
	err           error
	websocketConn *websocket.Conn
}

func New(name string, config *Config.WebsocketListener) (*WebsocketListener, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	listener := &WebsocketListener{
		name:       name,
		config:     config,
		status:     Status.Stopped,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}
	listener.httpServer = HTTPServer.New(listener.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: listener.config.TcpServerConfig,
		},
		nil, nil,
		map[string]http.HandlerFunc{
			listener.config.Pattern: listener.getHTTPWebsocketUpgradeHandler(),
		},
		nil,
	)
	simultaneousAccepts := []any{}
	for i := 0; i < listener.config.MaxSimultaneousAccepts; i++ {
		simultaneousAccepts = append(simultaneousAccepts, &acceptRequest{})
	}
	anySemaphore, err := Tools.NewGenericSemaphore(simultaneousAccepts)
	if err != nil {
		return nil, err
	}
	listener.anySemaphore = anySemaphore

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
