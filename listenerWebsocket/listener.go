package listenerWebsocket

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/httpServer"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type WebsocketListener struct {
	config *configs.WebsocketListener
	name   string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int
	stopChannel chan struct{}

	httpServer *httpServer.HTTPServer

	upgradeRequests chan (<-chan *upgraderResponse)
	timeout         *tools.Timeout
	mutex           sync.Mutex

	incomingMessageByteLimit uint64

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
	ClientsRejected atomic.Uint64
}

type upgraderResponse struct {
	err           error
	websocketConn *websocket.Conn
}

func New(name string, httpWrapperHandler httpServer.WrapperHandler, config *configs.WebsocketListener, incomingMessageByteLimit uint64, connectionLifetimeNs int64) (systemge.Listener[[]byte, systemge.Connection[[]byte]], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpListenerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	if config.Upgrader == nil {
		config.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}

	listener := &WebsocketListener{
		name:                     name,
		config:                   config,
		status:                   status.Stopped,
		instanceId:               tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
		upgradeRequests:          make(chan (<-chan *upgraderResponse)),
		incomingMessageByteLimit: incomingMessageByteLimit,
	}
	httpServer, err := httpServer.New(listener.name+"_httpServer",
		&configs.HTTPServer{
			TcpListenerConfig: listener.config.TcpListenerConfig,
		},
		httpWrapperHandler,
		map[string]http.HandlerFunc{
			listener.config.Pattern: listener.getHTTPWebsocketUpgradeHandler(),
		},
	)
	if err != nil {
		return nil, err
	}
	listener.httpServer = httpServer

	return listener, nil
}

func (listener *WebsocketListener) GetConnector() systemge.Connector[[]byte, systemge.Connection[[]byte]] {
	return &connector{
		tcpClientConfig: &configs.TcpClient{
			Port:    listener.config.TcpListenerConfig.Port,
			TlsCert: helpers.GetFileContent(listener.config.TcpListenerConfig.TlsCertPath),
			Domain:  listener.config.TcpListenerConfig.Domain,
		},
		incomingDataByteLimit: listener.incomingMessageByteLimit,
	}
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
