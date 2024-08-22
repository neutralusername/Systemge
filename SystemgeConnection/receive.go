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
			connection.messagesInProcessingChannel.Add(1)
			connection.waitGroup.Add(1)
			messageBytes, err := connection.receive()
			if err != nil {
				if connection.warningLogger != nil {
					connection.warningLogger.Log(Error.New("failed to receive message", err).Error())
				}
				connection.messagesInProcessingChannel.Add(-1)
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
			if err := connection.checkRateLimits(messageBytes); err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.messagesInProcessingChannel.Add(-1)
				connection.waitGroup.Done()
				if connection.errorLogger != nil {
					connection.errorLogger.Log(Error.New("failed to check rate limits for message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			message, err := Message.Deserialize(messageBytes, connection.GetName())
			if err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.messagesInProcessingChannel.Add(-1)
				connection.waitGroup.Done()
				if warningLogger := connection.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to deserialize message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			if err := connection.validateMessage(message); err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.messagesInProcessingChannel.Add(-1)
				connection.waitGroup.Done()
				if warningLogger := connection.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to validate message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			if infoLogger := connection.infoLogger; infoLogger != nil {
				infoLogger.Log("queueing message #" + Helpers.Uint64ToString(messageId) + " for processing")
			}
			if message.IsResponse() {
				if err := connection.AddSyncResponse(message); err != nil {
					connection.invalidMessagesReceived.Add(1)
					if warningLogger := connection.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("failed to add sync response for message #"+Helpers.Uint64ToString(messageId), err).Error())
					}
				}
				connection.messagesInProcessingChannel.Add(-1)
				connection.waitGroup.Done()
				continue
			} else {
				connection.processingChannel <- &messageInProcess{
					message: message,
					id:      messageId,
				}
			}
		}
	}
}

func (connection *SystemgeConnection) MessagesInProcessingChannel() int64 {
	return connection.messagesInProcessingChannel.Load()
}

func (connection *SystemgeConnection) GetNextMessage() (*Message.Message, error) {
	connection.processMutex.Lock()
	defer connection.processMutex.Unlock()
	if connection.processingLoopStopChannel != nil {
		return nil, Error.New("Processing loop already running", nil)
	}
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}
	select {
	case message := <-connection.processingChannel:
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Returned message # in GetNextMessage" + Helpers.Uint64ToString(message.id))
		}
		connection.waitGroup.Done()
		connection.messagesInProcessingChannel.Add(-1)
		return message.message, nil
	case <-timeout:
		return nil, Error.New("Timeout while waiting for message", nil)
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

func (connection *SystemgeConnection) StartProcessingLoopSequentially() error {
	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.processMutex.Unlock()
		return Error.New("Processing loop already running", nil)
	}
	connection.processingLoopStopChannel = make(chan bool)
	connection.processMutex.Unlock()
	go func() {
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Starting processing messages sequentially")
		}
		for {
			select {
			case message := <-connection.processingChannel:
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
				if err := connection.ProcessMessage(message.message); err != nil {
					if connection.warningLogger != nil {
						connection.warningLogger.Log(Error.New("Failed to process message #"+Helpers.Uint64ToString(message.id), err).Error())
					}
				} else {
					if infoLogger := connection.infoLogger; infoLogger != nil {
						infoLogger.Log("Processed message #" + Helpers.Uint64ToString(message.id))
					}
				}
				connection.messagesInProcessingChannel.Add(-1)
				connection.waitGroup.Done()
			case <-connection.processingLoopStopChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Stopping processing messages sequentially")
				}
				connection.processingLoopStopChannel = nil
				return
			}
		}
	}()
	return nil
}

func (connection *SystemgeConnection) StartProcessingLoopConcurrently() error {
	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.processMutex.Unlock()
		return Error.New("Processing loop already running", nil)
	}
	connection.processingLoopStopChannel = make(chan bool)
	connection.processMutex.Unlock()
	go func() {
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Starting processing messages concurrently")
		}
		for {
			select {
			case message := <-connection.processingChannel:
				go func() {
					if err := connection.ProcessMessage(message.message); err != nil {
						if connection.warningLogger != nil {
							connection.warningLogger.Log(Error.New("Failed to process message #"+Helpers.Uint64ToString(message.id), err).Error())
						}
					} else {
						if infoLogger := connection.infoLogger; infoLogger != nil {
							infoLogger.Log("Processed message #" + Helpers.Uint64ToString(message.id))
						}
					}
					connection.messagesInProcessingChannel.Add(-1)
					connection.waitGroup.Done()
				}()
			case <-connection.processingLoopStopChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Stopping processing messages concurrently")
				}
				connection.processingLoopStopChannel = nil
				return
			}
		}
	}()
	return nil
}

func (connection *SystemgeConnection) ProcessMessage(message *Message.Message) error {
	if connection.messageHandler != nil {
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
