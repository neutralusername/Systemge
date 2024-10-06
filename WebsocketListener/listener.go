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

	pool *Tools.GenericPool[*acceptRequest]

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
	ClientsRejected atomic.Uint64
}

type acceptRequest struct {
	upgraderResponseChannel chan *upgraderResponse
	timeoutMs               uint32
	triggered               sync.WaitGroup
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
	if config.MaxSimultaneousAccepts == 0 {
		config.MaxSimultaneousAccepts = 1
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
	pool, err := Tools.NewGenericPool(config.MaxSimultaneousAccepts, []*acceptRequest{}...)
	if err != nil {
		return nil, err
	}

	listener.pool = pool

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
