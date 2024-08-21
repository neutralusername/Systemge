package SystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *SystemgeConnection) receiveLoop() {
	if connection.infoLogger != nil {
		connection.infoLogger.Log("Started receiving messages")
	}

	for {
		select {
		case <-connection.closeChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopped receiving messages")
			}
			connection.Close()
			return
		default:
			connection.unprocessedMessages.Add(1)
			connection.waitGroup.Add(1)
			messageBytes, err := connection.receive()
			if err != nil {
				if connection.warningLogger != nil {
					connection.warningLogger.Log(Error.New("failed to receive message", err).Error())
				}
				connection.unprocessedMessages.Add(-1)
				connection.waitGroup.Done()
				if Tcp.IsConnectionClosed(err) {
					connection.Close()
					return
				}
				continue
			}
			connection.messageId++
			messageId := connection.messageId
			if infoLogger := connection.infoLogger; infoLogger != nil {
				infoLogger.Log("Received message #" + Helpers.Uint64ToString(messageId))
			}
			connection.processingChannel <- func() {
				if err := connection.processMessage(messageBytes, messageId); err != nil {
					if connection.warningLogger != nil {
						connection.warningLogger.Log(Error.New("Failed to process message #"+Helpers.Uint64ToString(messageId), err).Error())
					}
				}
				connection.unprocessedMessages.Add(-1)
				connection.waitGroup.Done()
			}
		}
	}
}

func (connection *SystemgeConnection) UnprocessedMessages() int64 {
	return connection.unprocessedMessages.Load()
}

func (connection *SystemgeConnection) ProcessNextMessage() error {
	connection.processMutex.Lock()
	defer connection.processMutex.Unlock()
	if connection.processingLoopStopChannel != nil {
		return Error.New("Processing loop already running", nil)
	}
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}
	select {
	case process := <-connection.processingChannel:
		if process == nil {
			return Error.New("Connection closed", nil)
		}
		process()
		return nil
	case <-timeout:
		return Error.New("Timeout while waiting for message", nil)
	}
}

func (connection *SystemgeConnection) StopProcessingLoop() error {
	connection.processMutex.Lock()
	defer connection.processMutex.Unlock()
	if connection.processingLoopStopChannel == nil {
		return Error.New("Processing loop not running", nil)
	}
	close(connection.processingLoopStopChannel)
	return nil
}

func (connection *SystemgeConnection) StartProcessingLoopSequentially() {
	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.processMutex.Unlock()
		return
	}
	connection.processingLoopStopChannel = make(chan bool)
	connection.processMutex.Unlock()

	if connection.infoLogger != nil {
		connection.infoLogger.Log("Starting processing messages sequentially")
	}
	for {
		select {
		case process := <-connection.processingChannel:
			if connection.config.ProcessingChannelSize > 0 && len(connection.processingChannel) >= connection.config.ProcessingChannelSize-1 {
				if connection.errorLogger != nil {
					connection.errorLogger.Log("Processing channel capacity reached")
				}
				if connection.mailer != nil {
					err := connection.mailer.Send(Tools.NewMail(nil, "error", Error.New("processing channel capacity reached", nil).Error()))
					if err != nil {
						if connection.errorLogger != nil {
							connection.errorLogger.Log(Error.New("failed sending mail", err).Error())
						}
					}
				}
			}
			process()
		case <-connection.processingLoopStopChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopping processing messages sequentially")
			}
			connection.processingLoopStopChannel = nil
			return
		}
	}
}

func (connection *SystemgeConnection) StartProcessingLoopConcurrently() {
	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.processMutex.Unlock()
		return
	}
	connection.processingLoopStopChannel = make(chan bool)
	connection.processMutex.Unlock()

	if connection.infoLogger != nil {
		connection.infoLogger.Log("Starting processing messages concurrently")
	}
	for {
		select {
		case process := <-connection.processingChannel:
			go process()
		case <-connection.processingLoopStopChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopping processing messages concurrently")
			}
			connection.processingLoopStopChannel = nil
			return
		}
	}
}

func (connection *SystemgeConnection) processMessage(messageBytes []byte, messageId uint64) error {
	if infoLogger := connection.infoLogger; infoLogger != nil {
		infoLogger.Log("Processing message #" + Helpers.Uint64ToString(messageId))
	}
	if err := connection.checkRateLimits(messageBytes); err != nil {
		return Error.New("rejected message due to rate limits", err)
	}
	message, err := Message.Deserialize(messageBytes, connection.GetName())
	if err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Error.New("failed to deserialize message", err)
	}
	if err := connection.validateMessage(message); err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Error.New("failed to validate message", err)
	}
	if message.IsResponse() {
		if err := connection.addSyncResponse(message); err != nil {
			connection.invalidMessagesReceived.Add(1)
			return Error.New("failed to add sync response message", err)
		}
	} else if connection.messageHandler != nil {
		if message.GetSyncTokenToken() == "" {
			connection.asyncMessagesReceived.Add(1)
			err := connection.messageHandler.HandleAsyncMessage(message)
			if err != nil {
				connection.invalidMessagesReceived.Add(1)
				return Error.New("failed to handle async message", err)
			}
		} else {
			connection.syncRequestsReceived.Add(1)
			if responsePayload, err := connection.messageHandler.HandleSyncRequest(message); err != nil {
				if err := connection.send(message.NewFailureResponse(err.Error()).Serialize()); err != nil {
					return Error.New("failed to send failure response", err)
				}
			} else {
				if err := connection.send(message.NewSuccessResponse(responsePayload).Serialize()); err != nil {
					return Error.New("failed to send success response", err)
				}
			}
		}
	} else {
		connection.invalidMessagesReceived.Add(1)
		return Error.New("no message handler available", nil)
	}
	if infoLogger := connection.infoLogger; infoLogger != nil {
		infoLogger.Log("Processed message #" + Helpers.Uint64ToString(messageId))
	}
	return nil
}

func (connection *SystemgeConnection) checkRateLimits(messageBytes []byte) error {
	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		connection.byteRateLimiterExceeded.Add(1)
		return Error.New("receiver rate limiter bytes exceeded", nil)
	}
	if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
		connection.messageRateLimiterExceeded.Add(1)
		return Error.New("receiver rate limiter messages exceeded", nil)
	}
	return nil
}

func (connection *SystemgeConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncTokenToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
