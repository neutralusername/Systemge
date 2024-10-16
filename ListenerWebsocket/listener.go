package ListenerWebsocket

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
	"github.com/neutralusername/Systemge/Systemge"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketListener struct {
	config *Config.WebsocketListener
	name   string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int

	acceptMutex sync.RWMutex

	stopChannel chan struct{}

	acceptRoutine *Tools.Routine

	httpServer *HTTPServer.HTTPServer

	upgradeRequests chan (<-chan *upgraderResponse)

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
	ClientsRejected atomic.Uint64
}

type upgraderResponse struct {
	err           error
	websocketConn *websocket.Conn
}

func New(name string, httpWrapperHandler HTTPServer.WrapperHandler, config *Config.WebsocketListener) (Systemge.SystemgeListener[[]byte], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	listener := &WebsocketListener{
		name:            name,
		config:          config,
		status:          Status.Stopped,
		instanceId:      Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		upgradeRequests: make(chan (<-chan *upgraderResponse)),
	}
	listener.httpServer = HTTPServer.New(listener.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: listener.config.TcpServerConfig,
		},
		httpWrapperHandler,
		map[string]http.HandlerFunc{
			listener.config.Pattern: listener.getHTTPWebsocketUpgradeHandler(),
		},
	)

	return listener, nil
}

func (listener *WebsocketListener) GetStopChannel() <-chan struct{} {
	return listener.stopChannel
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
