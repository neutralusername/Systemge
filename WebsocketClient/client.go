package WebsocketClient

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketClient struct {
	config        *Config.WebsocketClient
	websocketConn *websocket.Conn

	instanceId string

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool

	readHandler            Tools.ReadHandler[*WebsocketClient]
	readRoutineStopChannel chan struct{}
	readRoutineWaitGroup   sync.WaitGroup

	writeMutex sync.Mutex
	readMutex  sync.Mutex

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(config *Config.WebsocketClient, websocketConn *websocket.Conn) (*WebsocketClient, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if websocketConn == nil {
		return nil, errors.New("websocketConn is nil")
	}

	client := &WebsocketClient{
		config:        config,
		websocketConn: websocketConn,
		closeChannel:  make(chan bool),
		instanceId:    Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}
	websocketConn.SetReadLimit(int64(client.config.IncomingMessageByteLimit))

	return client, nil
}

func (client *WebsocketClient) GetStatus() int {
	client.closedMutex.Lock()
	defer client.closedMutex.Unlock()
	if client.closed {
		return Status.Stoped
	} else {
		return Status.Started
	}
}

func (client *WebsocketClient) GetInstanceId() string {
	return client.instanceId
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connclientction *WebsocketClient) GetCloseChannel() <-chan bool {
	return connclientction.closeChannel
}

func (client *WebsocketClient) GetAddress() string {
	return client.websocketConn.RemoteAddr().String()
}

func isWebsocketConnClosedErr(err error) bool {
	if err == nil {
		return false
	}

	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		return true
	}

	if websocket.IsUnexpectedCloseError(err) {
		return true
	}

	if errors.Is(err, websocket.ErrCloseSent) {
		return true
	}

	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
		return true
	}

	if nErr, ok := err.(net.Error); ok && !nErr.Timeout() {
		return true
	}

	return false
}
