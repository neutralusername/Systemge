package WebsocketClient

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketClient struct {
	name          string
	config        *Config.WebsocketClient
	websocketConn *websocket.Conn

	instanceId string

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool

	sendMutex sync.Mutex
	readMutex sync.Mutex

	eventHandler Event.Handler

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(config *Config.WebsocketClient, websocketConn *websocket.Conn, eventHandler Event.Handler) (*WebsocketClient, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if websocketConn == nil {
		return nil, errors.New("websocketConn is nil")
	}

	connection := &WebsocketClient{
		config:        config,
		websocketConn: websocketConn,
		closeChannel:  make(chan bool),
		eventHandler:  eventHandler,
		instanceId:    Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}
	websocketConn.SetReadLimit(int64(connection.config.IncomingMessageByteLimit))

	return connection, nil
}

func (connection *WebsocketClient) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.Stopped
	} else {
		return Status.Started
	}
}

func (connection *WebsocketClient) SetName(name string) {
	connection.name = name
}

func (connection *WebsocketClient) GetName() string {
	return connection.name
}

func (connection *WebsocketClient) GetInstanceId() string {
	return connection.instanceId
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *WebsocketClient) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *WebsocketClient) GetAddress() string {
	return connection.websocketConn.RemoteAddr().String()
}

func (connection *WebsocketClient) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(connection.GetContext())
	if connection.eventHandler != nil {
		connection.eventHandler(event)
	}
	return event
}
func (connection *WebsocketClient) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.WebsocketClient,
		Event.ServiceName:       connection.name,
		Event.ServiceStatus:     Status.ToString(connection.GetStatus()),
		Event.ServiceInstanceId: connection.instanceId,
		Event.Address:           connection.GetAddress(),
		//	Event.Function:          Event.GetCallerFuncName(2),
	}
}
