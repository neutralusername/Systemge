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
	"github.com/neutralusername/Systemge/WebsocketClient"
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

	acceptHandler            Tools.AcceptHandler[*WebsocketClient.WebsocketClient]
	acceptRoutineStopChannel chan struct{}
	acceptRoutineWaitGroup   sync.WaitGroup
	acceptRoutineSemaphore   *Tools.Semaphore[struct{}]

	waitgroup sync.WaitGroup

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

func New(name string, config *Config.WebsocketListener, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, ipRateLimiter *Tools.IpRateLimiter) (*WebsocketListener, error) {
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
		whitelist, blacklist, ipRateLimiter,
		map[string]http.HandlerFunc{
			listener.config.Pattern: listener.getHTTPWebsocketUpgradeHandler(),
		},
		nil,
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
