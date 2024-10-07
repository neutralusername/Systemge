package WebsocketClient

import (
	"errors"
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

	state map[string]any

	sendMutex sync.Mutex
	readMutex sync.Mutex

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
		state:         make(map[string]any),
		instanceId:    Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}
	websocketConn.SetReadLimit(int64(client.config.IncomingMessageByteLimit))

	return client, nil
}

func (client *WebsocketClient) SetState(key string, value any) {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	client.state[key] = value
}

func (client *WebsocketClient) GetState(key string) any {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()
	return client.state[key]
}

func (client *WebsocketClient) RemoveState(key string) {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()
	delete(client.state, key)
}

func (client *WebsocketClient) GetStatus() int {
	client.closedMutex.Lock()
	defer client.closedMutex.Unlock()
	if client.closed {
		return Status.Stopped
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
