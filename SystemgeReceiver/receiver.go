package SystemgeReceiver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeReceiver struct {
	status      int
	statusMutex sync.Mutex

	config         *Config.SystemgeReceiver
	connection     *SystemgeConnection.SystemgeConnection
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler

	rateLimiterBytes    *Tools.TokenBucketRateLimiter
	rateLimiterMessages *Tools.TokenBucketRateLimiter

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer

	messageChannel chan func()
	waitGroup      sync.WaitGroup

	// metrics

	asyncMessagesReceived   atomic.Uint32
	syncRequestsReceived    atomic.Uint32
	invalidMessagesReceived atomic.Uint32
	syncResponsesReceived   atomic.Uint32

	messageRateLimiterExceeded atomic.Uint32
	byteRateLimiterExceeded    atomic.Uint32
}

func New(config *Config.SystemgeReceiver, connection *SystemgeConnection.SystemgeConnection, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeReceiver {
	if config == nil {
		panic("config is nil")
	}
	if connection == nil {
		panic("connection is nil")
	}
	receiver := &SystemgeReceiver{
		config:         config,
		connection:     connection,
		messageHandler: messageHandler,
	}
	if config.RateLimiterBytes != nil {
		receiver.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(config.RateLimiterBytes)
	}
	if config.RateLimiterMessages != nil {
		receiver.rateLimiterMessages = Tools.NewTokenBucketRateLimiter(config.RateLimiterMessages)
	}
	if config.InfoLoggerPath != "" {
		receiver.infoLogger = Tools.NewLogger("[Info: \""+connection.GetName()+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		receiver.warningLogger = Tools.NewLogger("[Warning: \""+connection.GetName()+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		receiver.errorLogger = Tools.NewLogger("[Error: \""+connection.GetName()+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		receiver.mailer = Tools.NewMailer(config.MailerConfig)
	}
	return receiver
}

func (receiver *SystemgeReceiver) Start() error {
	receiver.statusMutex.Lock()
	defer receiver.statusMutex.Unlock()
	if receiver.status != Status.STOPPED {
		return Error.New("receiver already started", nil)
	}
	receiver.messageChannel = make(chan func())
	receiver.status = Status.STARTED
	if receiver.config.ProcessSequentially {
		go receiver.processingLoopSequentially()
	} else {
		go receiver.processingLoopConcurrently()
	}
	go receiver.receiveLoop()
	return nil
}

func (receiver *SystemgeReceiver) Stop() error {
	receiver.statusMutex.Lock()
	defer receiver.statusMutex.Unlock()
	if receiver.status != Status.STARTED {
		return Error.New("receiver already stopped", nil)
	}
	close(receiver.messageChannel)
	receiver.waitGroup.Wait()
	receiver.status = Status.STOPPED
	return nil
}
