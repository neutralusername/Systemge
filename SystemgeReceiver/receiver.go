package SystemgeReceiver

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
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

	errorLogger    *Tools.Logger
	warningLogger  *Tools.Logger
	messageChannel chan func()
	running        bool
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
	return receiver
}

func (receiver *SystemgeReceiver) Start() error {
	receiver.statusMutex.Lock()
	defer receiver.statusMutex.Unlock()
	if receiver.status != Status.STOPPED {
		return Error.New("receiver already started", nil)
	}
	receiver.running = true
	receiver.messageChannel = make(chan func())
	go receiver.receiveLoop()
	go receiver.processingLoop()
	receiver.status = Status.STARTED
	return nil
}

func (receiver *SystemgeReceiver) Stop() error {
	receiver.statusMutex.Lock()
	defer receiver.statusMutex.Unlock()
	if receiver.status != Status.STARTED {
		return Error.New("receiver already stopped", nil)
	}
	receiver.running = false
	close(receiver.messageChannel)
	receiver.waitGroup.Wait()
	receiver.status = Status.STOPPED
	return nil
}

func (receiver *SystemgeReceiver) processingLoop() {
	for {
		select {
		case process := <-receiver.messageChannel:
			if process != nil {
				return
			}
			process()
		}
	}
}

func (receiver *SystemgeReceiver) receiveLoop() {
	for receiver.running {
		messageBytes, err := receiver.connection.ReceiveMessage()
		if err != nil {
			if receiver.errorLogger != nil {
				receiver.errorLogger.Log(Error.New("failed to receive message", err).Error())
			}
			continue
		}
		receiver.waitGroup.Add(1)
		receiver.messageChannel <- func() {
			err := receiver.processMessage(receiver.connection, messageBytes)
			if err != nil {
				if receiver.warningLogger != nil {
					receiver.warningLogger.Log(Error.New("failed to process message", err).Error())
				}
			}
		}
	}
}

func (receiver *SystemgeReceiver) processMessage(clientConnection *SystemgeConnection.SystemgeConnection, messageBytes []byte) error {
	defer receiver.waitGroup.Done()
	if err := receiver.checkRateLimits(clientConnection, messageBytes); err != nil {
		return Error.New("rejected message due to rate limits", err)
	}
	message, err := Message.Deserialize(messageBytes, clientConnection.GetName())
	if err != nil {
		receiver.invalidMessagesReceived.Add(1)
		return Error.New("failed to deserialize message", err)
	}
	if err := receiver.validateMessage(message); err != nil {
		receiver.invalidMessagesReceived.Add(1)
		return Error.New("failed to validate message", err)
	}
	if message.GetSyncTokenToken() == "" {
		receiver.asyncMessagesReceived.Add(1)
		receiver.messageHandler.HandleAsyncMessage(message)
	} else if message.IsResponse() {
		if err := receiver.connection.AddSyncResponse(message); err != nil {
			receiver.invalidMessagesReceived.Add(1)
			return Error.New("failed to add sync response message", err)
		} else {
			receiver.syncResponsesReceived.Add(1)
		}
	} else {
		receiver.syncRequestsReceived.Add(1)
		if responsePayload, err := receiver.messageHandler.HandleSyncRequest(message); err != nil {
			if err := receiver.connection.SendMessage(message.NewFailureResponse(err.Error()).Serialize()); err != nil {
				return Error.New("failed to send failure response", err)
			}
		} else {
			if err := receiver.connection.SendMessage(message.NewSuccessResponse(responsePayload).Serialize()); err != nil {
				return Error.New("failed to send success response", err)
			}
		}
	}
	return nil
}

func (receiver *SystemgeReceiver) checkRateLimits(clientConnection *SystemgeConnection.SystemgeConnection, messageBytes []byte) error {
	if receiver.rateLimiterBytes != nil && !receiver.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		receiver.byteRateLimiterExceeded.Add(1)
		return Error.New("client connection rate limiter bytes exceeded", nil)
	}
	if receiver.rateLimiterMessages != nil && !receiver.rateLimiterMessages.Consume(1) {
		receiver.messageRateLimiterExceeded.Add(1)
		return Error.New("client connection rate limiter messages exceeded", nil)
	}
	return nil
}

func (receiver *SystemgeReceiver) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := receiver.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncTokenToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := receiver.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := receiver.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
