package SystemgeConnection

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/Tools"
)

type messageInProcess struct {
	message *Message.Message
	id      uint64
}

type SystemgeConnection struct {
	name       string
	config     *Config.SystemgeConnection
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

	tcpBuffer []byte

	syncResponseChannels map[string]chan *Message.Message
	syncMutex            sync.Mutex

	closeChannel chan bool

	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler

	processMutex                sync.Mutex
	processingChannel           chan *messageInProcess
	processingLoopStopChannel   chan bool
	messagesInProcessingChannel atomic.Int64
	waitGroup                   sync.WaitGroup

	rateLimiterBytes    *Tools.TokenBucketRateLimiter
	rateLimiterMessages *Tools.TokenBucketRateLimiter

	messageId uint64

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	asyncMessagesSent atomic.Uint64
	syncRequestsSent  atomic.Uint64

	syncSuccessResponsesReceived atomic.Uint64
	syncFailureResponsesReceived atomic.Uint64
	noSyncResponseReceived       atomic.Uint64

	asyncMessagesReceived   atomic.Uint64
	syncRequestsReceived    atomic.Uint64
	invalidMessagesReceived atomic.Uint64

	messageRateLimiterExceeded atomic.Uint64
	byteRateLimiterExceeded    atomic.Uint64
}

func New(config *Config.SystemgeConnection, netConn net.Conn, name string, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeConnection {
	connection := &SystemgeConnection{
		name:                 name,
		config:               config,
		netConn:              netConn,
		randomizer:           Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:         make(chan bool),
		syncResponseChannels: make(map[string]chan *Message.Message),
		processingChannel:    make(chan *messageInProcess, config.ProcessingChannelSize),
		messageHandler:       messageHandler,
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
	go connection.receiveLoop()
	return connection
}

// if processing loop is running, this function will block until all messages are processed
func (connection *SystemgeConnection) Close() {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return
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

	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.waitGroup.Wait()
		close(connection.processingLoopStopChannel)
		connection.processingLoopStopChannel = nil
	}
	connection.processMutex.Unlock()
}

func (connection *SystemgeConnection) IsClosed() bool {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	return connection.closed
}

func (connection *SystemgeConnection) GetName() string {
	return connection.name
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *SystemgeConnection) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *SystemgeConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}
