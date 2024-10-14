package TcpSystemgeConnection

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpSystemgeConnection struct {
	name       string
	config     *Config.TcpSystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	readMutex  sync.RWMutex
	writeMutex sync.Mutex

	readRoutine *Tools.Routine

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool
	waitGroup    sync.WaitGroup

	messageReceiver *Tcp.BufferedMessageReader

	// metrics
	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeConnection, netConn net.Conn, messageReceiver *Tcp.BufferedMessageReader, eventHandler Event.Handler) (*TcpSystemgeConnection, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if netConn == nil {
		return nil, errors.New("netConn is nil")
	}
	if messageReceiver == nil {
		return nil, errors.New("messageReceiver is nil")
	}

	connection := &TcpSystemgeConnection{
		name:            name,
		config:          config,
		netConn:         netConn,
		messageReceiver: messageReceiver,
		randomizer:      Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:    make(chan bool),
	}
	if config.TcpBufferBytes <= 0 {
		config.TcpBufferBytes = 1024 * 4
	}

	connection.waitGroup.Add(1)
	go connection.receptionRoutine()
	if config.HeartbeatIntervalMs > 0 {
		connection.waitGroup.Add(1)
		go connection.heartbeatLoop()
	}
	return connection, nil
}

func (connection *TcpSystemgeConnection) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("connection already closing")
	}
	defer connection.closedMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.ServiceStoping,
		"closing tcpSystemgeConnection",
		Event.Skip,
		Event.Skip,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.closed {
		connection.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"service tcpSystemgeConnection already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
		return errors.New("connection already closed")
	}

	connection.closed = true
	connection.netConn.Close()
	close(connection.closeChannel)
	connection.waitGroup.Wait()

	if connection.rateLimiterBytes != nil {
		connection.rateLimiterBytes.Close()
		connection.rateLimiterBytes = nil
	}
	if connection.rateLimiterMessages != nil {
		connection.rateLimiterMessages.Close()
		connection.rateLimiterMessages = nil
	}
	close(connection.messageChannel)

	connection.onEvent(Event.NewInfoNoOption(
		Event.ServiceStoped,
		"connection closed",
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	))

	return nil
}

func (connection *TcpSystemgeConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.Stopped
	} else {
		return Status.Started
	}
}

func (connection *TcpSystemgeConnection) GetName() string {
	return connection.name
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *TcpSystemgeConnection) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *TcpSystemgeConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}

func (connection *TcpSystemgeConnection) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(connection.GetServerContext())
	if connection.eventHandler != nil {
		connection.eventHandler(event)
	}
	return event
}
func (connection *TcpSystemgeConnection) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.TcpSystemgeListener,
		Event.ServiceName:   connection.name,
		Event.Address:       connection.GetAddress(),
		Event.ServiceStatus: Status.ToString(connection.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
}
