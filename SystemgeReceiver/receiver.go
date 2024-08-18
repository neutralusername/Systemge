package SystemgeReceiver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeReceiver struct {
	config         *Config.SystemgeReceiver
	connection     *SystemgeConnection.SystemgeConnection
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler

	rateLimiterBytes    *Tools.TokenBucketRateLimiter
	rateLimiterMessages *Tools.TokenBucketRateLimiter

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer

	processingChannel chan func()
	waitGroup         sync.WaitGroup

	closed      bool
	closedMutex sync.Mutex

	messageId uint32

	// metrics

	asyncMessagesReceived   atomic.Uint32
	syncRequestsReceived    atomic.Uint32
	invalidMessagesReceived atomic.Uint32
	syncResponsesReceived   atomic.Uint32

	messageRateLimiterExceeded atomic.Uint32
	byteRateLimiterExceeded    atomic.Uint32
}

func New(connection *SystemgeConnection.SystemgeConnection, config *Config.SystemgeReceiver, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeReceiver {
	if config == nil {
		panic("config is nil")
	}
	if connection == nil {
		panic("connection is nil")
	}
	if messageHandler == nil {
		panic("messageHandler is nil")
	}
	receiver := &SystemgeReceiver{
		config:         config,
		connection:     connection,
		messageHandler: messageHandler,
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
	if receiver.config.RateLimiterBytes != nil {
		receiver.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(receiver.config.RateLimiterBytes)
	}
	if receiver.config.RateLimiterMessages != nil {
		receiver.rateLimiterMessages = Tools.NewTokenBucketRateLimiter(receiver.config.RateLimiterMessages)
	}
	receiver.processingChannel = make(chan func(), receiver.config.ProcessingChannelSize)
	if receiver.config.ProcessSequentially {
		go receiver.processingLoopSequentially(receiver.processingChannel)
	} else {
		go receiver.processingLoopConcurrently(receiver.processingChannel)
	}
	go receiver.receive(receiver.processingChannel)
	return receiver
}

func (receiver *SystemgeReceiver) Close() {
	receiver.closedMutex.Lock()
	defer receiver.closedMutex.Unlock()
	if receiver.closed {
		return
	}
	receiver.closed = true
	if receiver.rateLimiterBytes != nil {
		receiver.rateLimiterBytes.Stop()
		receiver.rateLimiterBytes = nil
	}
	if receiver.rateLimiterMessages != nil {
		receiver.rateLimiterMessages.Stop()
		receiver.rateLimiterMessages = nil
	}
	processingChannel := receiver.processingChannel
	receiver.processingChannel = nil
	receiver.waitGroup.Wait()
	close(processingChannel)
}
