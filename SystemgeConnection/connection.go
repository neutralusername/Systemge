package SystemgeConnection

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

	syncRequests map[string]*syncRequestStruct
	syncMutex    sync.Mutex

	closeChannel chan bool

	processMutex              sync.Mutex
	processingChannel         chan *messageInProcess
	processingLoopStopChannel chan bool
	unprocessedMessages       atomic.Int64

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
	invalidSyncResponsesReceived atomic.Uint64

	invalidMessagesReceived atomic.Uint64
	validMessagesReceived   atomic.Uint64

	messageRateLimiterExceeded atomic.Uint64
	byteRateLimiterExceeded    atomic.Uint64
}

func New(name string, config *Config.SystemgeConnection, netConn net.Conn) *SystemgeConnection {
	connection := &SystemgeConnection{
		name:              name,
		config:            config,
		netConn:           netConn,
		randomizer:        Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:      make(chan bool),
		syncRequests:      make(map[string]*syncRequestStruct),
		processingChannel: make(chan *messageInProcess, config.ProcessingChannelCapacity),
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
	go connection.receiveLoop()
	return connection
}

func (connection *SystemgeConnection) Close() error {
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
	return nil
}

func (connection *SystemgeConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.STOPPED
	} else {
		return Status.STARTED
	}
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
