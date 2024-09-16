package TcpSystemgeConnection

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpConnection struct {
	name       string
	config     *Config.TcpSystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

	closed      bool
	closedMutex sync.Mutex

	messageReceiver *Tools.MessageReceiver

	syncRequests map[string]*syncRequestStruct
	syncMutex    sync.Mutex

	closeChannel chan bool

	messageMutex            sync.Mutex
	messageChannel          chan *Message.Message
	messageChannelSemaphore *Tools.Semaphore
	receiveLoopStopChannel  chan bool

	rateLimiterBytes    *Tools.TokenBucketRateLimiter
	rateLimiterMessages *Tools.TokenBucketRateLimiter

	// metrics
	bytesSent atomic.Uint64

	asyncMessagesSent atomic.Uint64
	syncRequestsSent  atomic.Uint64

	syncSuccessResponsesReceived atomic.Uint64
	syncFailureResponsesReceived atomic.Uint64
	noSyncResponseReceived       atomic.Uint64
	invalidSyncResponsesReceived atomic.Uint64

	invalidMessagesReceived atomic.Uint64
	validMessagesReceived   atomic.Uint64

	messageRateLimiterExceeded atomic.Uint64
	byteRateLimiterExceeded    atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeConnection, netConn net.Conn, messageReceiver *Tools.MessageReceiver) *TcpConnection {
	connection := &TcpConnection{
		name:                    name,
		config:                  config,
		netConn:                 netConn,
		messageReceiver:         messageReceiver,
		randomizer:              Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:            make(chan bool),
		syncRequests:            make(map[string]*syncRequestStruct),
		messageChannel:          make(chan *Message.Message, config.ProcessingChannelCapacity+1), // +1 so that the receive loop is never blocking while adding a message to the processing channel
		messageChannelSemaphore: Tools.NewSemaphore(config.ProcessingChannelCapacity+1, config.ProcessingChannelCapacity+1),
		receiveLoopStopChannel:  make(chan bool),
	}
	if config.InfoLoggerPath != "" {
		connection.infoLogger = Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		connection.warningLogger = Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		connection.errorLogger = Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		connection.mailer = Tools.NewMailer(config.MailerConfig)
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
	go connection.addMessageToProcessingChannelLoop()
	if config.HeartbeatIntervalMs > 0 {
		go connection.heartbeatLoop()
	}
	return connection
}

func (connection *TcpConnection) Close() error {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return Error.New("Connection already closed", nil)
	}

	connection.closed = true
	close(connection.closeChannel)
	connection.netConn.Close()

	if connection.rateLimiterBytes != nil {
		connection.rateLimiterBytes.Close()
		connection.rateLimiterBytes = nil
	}
	if connection.rateLimiterMessages != nil {
		connection.rateLimiterMessages.Close()
		connection.rateLimiterMessages = nil
	}
	<-connection.receiveLoopStopChannel
	close(connection.messageChannel)
	return nil
}

func (connection *TcpConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.STOPPED
	} else {
		return Status.STARTED
	}
}

func (connection *TcpConnection) GetName() string {
	return connection.name
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *TcpConnection) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *TcpConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}
