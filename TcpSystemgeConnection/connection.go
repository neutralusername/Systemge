package TcpSystemgeConnection

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpSystemgeConnection struct {
	name       string
	config     *Config.TcpSystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	sendMutex sync.Mutex

	closed      bool
	closedMutex sync.Mutex

	messageReceiver *Tcp.BufferedMessageReceiver

	syncRequests map[string]*syncRequestStruct
	syncMutex    sync.Mutex

	closeChannel chan bool
	waitGroup    sync.WaitGroup

	eventHandler Event.Handler

	messageHandlingLoopStopChannel chan<- bool
	messageMutex                   sync.Mutex
	messageChannel                 chan *Message.Message
	messageChannelSemaphore        *Tools.Semaphore

	rateLimiterBytes    *Tools.TokenBucketRateLimiter
	rateLimiterMessages *Tools.TokenBucketRateLimiter

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	asyncMessagesSent atomic.Uint64
	syncRequestsSent  atomic.Uint64
	syncResponsesSent atomic.Uint64

	messagesReceived        atomic.Uint64
	invalidMessagesReceived atomic.Uint64
	rejectedMessages        atomic.Uint64

	syncSuccessResponsesReceived atomic.Uint64
	syncFailureResponsesReceived atomic.Uint64
	noSyncResponseReceived       atomic.Uint64
	invalidSyncResponsesReceived atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeConnection, netConn net.Conn, messageReceiver *Tcp.BufferedMessageReceiver, eventHandler Event.Handler) *TcpSystemgeConnection {
	connection := &TcpSystemgeConnection{
		name:                    name,
		config:                  config,
		netConn:                 netConn,
		messageReceiver:         messageReceiver,
		randomizer:              Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:            make(chan bool),
		syncRequests:            make(map[string]*syncRequestStruct),
		messageChannel:          make(chan *Message.Message, config.ProcessingChannelCapacity+1), // +1 so that the receive loop is never blocking while adding a message to the processing channel
		messageChannelSemaphore: Tools.NewSemaphore(config.ProcessingChannelCapacity+1, config.ProcessingChannelCapacity+1),
		eventHandler:            eventHandler,
	}
	if config.RateLimiterBytes != nil {
		connection.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(config.RateLimiterBytes)
	}
	if config.RateLimiterMessages != nil {
		connection.rateLimiterMessages = Tools.NewTokenBucketRateLimiter(config.RateLimiterMessages)
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
	return connection
}

func (connection *TcpSystemgeConnection) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("connection already closing")
	}
	defer connection.closedMutex.Unlock()

	if event := connection.eventHandler(Event.NewInfo(
		Event.StoppingService,
		"closing tcpSystemgeConnection",
		Event.Skip,
		Event.Skip,
		Event.Continue,
		Event.Context{},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.closed {
		connection.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"service tcpSystemgeConnection already started",
			Event.Context{
				Event.Circumstance: Event.Start,
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
		Event.ServiceStopped,
		"connection closed",
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	))

	return nil
}

func (connection *TcpSystemgeConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.Stoped
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

func (server *TcpSystemgeConnection) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler == nil {
		return event
	}
	return server.eventHandler(event)
}
func (server *TcpSystemgeConnection) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.TcpSystemgeListener,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
}

func (connection *TcpSystemgeConnection) GetIp() string {
	return connection.netConn.RemoteAddr().String()
}
