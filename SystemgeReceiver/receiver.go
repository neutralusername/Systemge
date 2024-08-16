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

	processingChannel chan func()
	waitGroup         sync.WaitGroup

	messageId uint32

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
	if messageHandler == nil {
		panic("messageHandler is nil")
	}
	receiver := &SystemgeReceiver{
		config:         config,
		connection:     connection,
		messageHandler: messageHandler,
	}
	if config.InfoLoggerPath != "" {
		receiver.infoLogger = Tools.NewLogger("[Info: \""+receiver.GetName()+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		receiver.warningLogger = Tools.NewLogger("[Warning: \""+receiver.GetName()+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		receiver.errorLogger = Tools.NewLogger("[Error: \""+receiver.GetName()+"\"] ", config.ErrorLoggerPath)
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
	receiver.status = Status.PENDING
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Starting receiver")
	}
	if receiver.config.RateLimiterBytes != nil {
		receiver.rateLimiterBytes = Tools.NewTokenBucketRateLimiter(receiver.config.RateLimiterBytes)
	}
	if receiver.config.RateLimiterMessages != nil {
		receiver.rateLimiterMessages = Tools.NewTokenBucketRateLimiter(receiver.config.RateLimiterMessages)
	}
	receiver.processingChannel = make(chan func(), receiver.config.ProcessingChannelSize)
	if receiver.config.ProcessSequentially {
		go receiver.processingLoopSequentially()
	} else {
		go receiver.processingLoopConcurrently()
	}
	go receiver.receive(receiver.processingChannel)
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Receiver started")
	}
	receiver.status = Status.STARTED
	return nil
}

func (receiver *SystemgeReceiver) Stop() error {
	receiver.statusMutex.Lock()
	defer receiver.statusMutex.Unlock()
	if receiver.status != Status.STARTED {
		return Error.New("receiver already stopped", nil)
	}
	receiver.status = Status.PENDING
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Stopping receiver")
	}
	if receiver.rateLimiterBytes != nil {
		receiver.rateLimiterBytes.Stop()
		receiver.rateLimiterBytes = nil
	}
	if receiver.rateLimiterMessages != nil {
		receiver.rateLimiterMessages.Stop()
		receiver.rateLimiterMessages = nil
	}
	messageChannel := receiver.processingChannel
	receiver.processingChannel = nil
	receiver.waitGroup.Wait()
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Receiver stopped")
	}
	close(messageChannel)
	receiver.status = Status.STOPPED
	return nil
}

func (receiver *SystemgeReceiver) GetStatus() int {
	return receiver.status
}

func (receiver *SystemgeReceiver) GetName() string {
	return receiver.config.Name
}
