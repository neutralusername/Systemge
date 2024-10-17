package listenerWebsocket

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/httpServer"
	"github.com/neutralusername/Systemge/status"
	"github.com/neutralusername/Systemge/systemge"
	"github.com/neutralusername/Systemge/tools"
)

type WebsocketListener struct {
	config *Config.WebsocketListener
	name   string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int
	stopChannel chan struct{}

	httpServer *httpServer.HTTPServer

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

func New(name string, httpWrapperHandler httpServer.WrapperHandler, config *Config.WebsocketListener) (systemge.Listener[[]byte, systemge.Connection[[]byte]], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	listener := &WebsocketListener{
		name:            name,
		config:          config,
		status:          status.Stopped,
		instanceId:      tools.GenerateRandomString(Constants.InstanceIdLength, tools.ALPHA_NUMERIC),
		upgradeRequests: make(chan (<-chan *upgraderResponse)),
	}
	listener.httpServer = httpServer.New(listener.name+"_httpServer",
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
