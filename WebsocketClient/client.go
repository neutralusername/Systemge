package WebsocketClient

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

type WebsocketClient struct {
	name          string
	config        *Config.WebsocketClient
	websocketConn *websocket.Conn

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool

	sendMutex sync.Mutex

	eventHandler Event.Handler

	// metrics

	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	messagesSent     atomic.Uint64
	messagesReceived atomic.Uint64
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
	}
	websocketConn.SetReadLimit(int64(connection.config.IncomingMessageByteLimit))

	return connection, nil
}

func (connection *WebsocketClient) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("websocketClient already closing")
	}
	defer connection.closedMutex.Unlock()

	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ServiceStoping,
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("close canceled")
		}
	}

	if connection.closed {
		if connection.eventHandler != nil {
			connection.onEvent(Event.New(
				Event.ServiceAlreadyStarted,
				Event.Context{
					Event.Circumstance: Event.ServiceStop,
				},
				Event.Cancel,
			))
		}
		return errors.New("websocketClient already closed")
	}

	connection.closed = true
	connection.websocketConn.Close()
	close(connection.closeChannel)

	if connection.eventHandler != nil {
		connection.onEvent(Event.New(
			Event.ServiceStoped,
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
			Event.Continue,
		))
	}

	return nil
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
		Event.ServiceType:   Event.WebsocketClient,
		Event.ServiceName:   connection.name,
		Event.Address:       connection.GetAddress(),
		Event.ServiceStatus: Status.ToString(connection.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
}
