package SystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
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
			messageBytes, err := connection.receive()
			if err != nil {
				if connection.warningLogger != nil {
					connection.warningLogger.Log(Error.New("failed to receive message", err).Error())
				}
				connection.unprocessedMessages.Add(-1)
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
			if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
				connection.byteRateLimiterExceeded.Add(1)
				connection.unprocessedMessages.Add(-1)
				if connection.warningLogger != nil {
					connection.warningLogger.Log("Byte rate limiter exceeded for message #" + Helpers.Uint64ToString(messageId))
				}
				continue
			}
			if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
				connection.messageRateLimiterExceeded.Add(1)
				connection.unprocessedMessages.Add(-1)
				if connection.warningLogger != nil {
					connection.warningLogger.Log("Message rate limiter exceeded for message #" + Helpers.Uint64ToString(messageId))
				}
				continue
			}
			message, err := Message.Deserialize(messageBytes, connection.GetName())
			if err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.unprocessedMessages.Add(-1)
				if warningLogger := connection.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to deserialize message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			if err := connection.validateMessage(message); err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.unprocessedMessages.Add(-1)
				if warningLogger := connection.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to validate message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			if infoLogger := connection.infoLogger; infoLogger != nil {
				infoLogger.Log("queueing message #" + Helpers.Uint64ToString(messageId) + " for processing")
			}
			if message.IsResponse() {
				if err := connection.addSyncResponse(message); err != nil {
					connection.invalidSyncResponsesReceived.Add(1)
					if warningLogger := connection.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("failed to add sync response for message #"+Helpers.Uint64ToString(messageId), err).Error())
					}
				}
				connection.unprocessedMessages.Add(-1)
				continue
			} else {
				connection.validMessagesReceived.Add(1)
				if connection.config.ProcessingChannelCapacity > 0 && len(connection.processingChannel) == cap(connection.processingChannel) {
					if connection.warningLogger != nil {
						connection.warningLogger.Log("Processing channel capacity reached for message #" + Helpers.Uint64ToString(messageId))
					}
				}
				connection.processingChannel <- &messageInProcess{
					message: message,
					id:      messageId,
				}
			}
		}
	}
}

func (connection *SystemgeConnection) UnprocessedMessagesCount() int64 {
	return connection.unprocessedMessages.Load()
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
		connection.unprocessedMessages.Add(-1)
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
	connection.processingLoopStopChannel = nil
	return nil
}

func (connection *SystemgeConnection) StartProcessingLoopSequentially(messageHandler SystemgeMessageHandler.MessageHandler) error {
	if messageHandler == nil {
		return Error.New("No message handler set", nil)
	}
	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.processMutex.Unlock()
		return Error.New("Processing loop already running", nil)
	}
	processingLoopStopChannel := make(chan bool)
	connection.processingLoopStopChannel = processingLoopStopChannel
	connection.processMutex.Unlock()
	go func() {
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Starting processing messages sequentially")
		}
		for {
			select {
			case message := <-connection.processingChannel:
				if err := connection.ProcessMessage(message.message, messageHandler); err != nil {
					if connection.errorLogger != nil {
						connection.errorLogger.Log(Error.New("Failed to process message #"+Helpers.Uint64ToString(message.id), err).Error())
					}
					if connection.mailer != nil {
						err := connection.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to process message #"+Helpers.Uint64ToString(message.id), err).Error()))
						if err != nil {
							if connection.errorLogger != nil {
								connection.errorLogger.Log(Error.New("Failed to send mail", err).Error())
							}
						}
					}
				} else {
					if infoLogger := connection.infoLogger; infoLogger != nil {
						infoLogger.Log("Processed message #" + Helpers.Uint64ToString(message.id))
					}
				}
				connection.unprocessedMessages.Add(-1)
			case <-processingLoopStopChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Stopping processing messages sequentially")
				}
				return
			}
		}
	}()
	return nil
}

func (connection *SystemgeConnection) StartProcessingLoopConcurrently(messageHandler SystemgeMessageHandler.MessageHandler) error {
	connection.processMutex.Lock()
	if connection.processingLoopStopChannel != nil {
		connection.processMutex.Unlock()
		return Error.New("Processing loop already running", nil)
	}
	processingLoopStopChannel := make(chan bool)
	connection.processingLoopStopChannel = processingLoopStopChannel
	connection.processMutex.Unlock()
	go func() {
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Starting processing messages concurrently")
		}
		for {
			select {
			case message := <-connection.processingChannel:
				go func() {
					if err := connection.ProcessMessage(message.message, messageHandler); err != nil {
						if connection.errorLogger != nil {
							connection.errorLogger.Log(Error.New("Failed to process message #"+Helpers.Uint64ToString(message.id), err).Error())
						}
						if connection.mailer != nil {
							err := connection.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to process message #"+Helpers.Uint64ToString(message.id), err).Error()))
							if err != nil {
								if connection.errorLogger != nil {
									connection.errorLogger.Log(Error.New("Failed to send mail", err).Error())
								}
							}
						}
					} else {
						if infoLogger := connection.infoLogger; infoLogger != nil {
							infoLogger.Log("Processed message #" + Helpers.Uint64ToString(message.id))
						}
					}
					connection.unprocessedMessages.Add(-1)
				}()
			case <-processingLoopStopChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Stopping processing messages concurrently")
				}
				return
			}
		}
	}()
	return nil
}

func (connection *SystemgeConnection) ProcessMessage(message *Message.Message, messageHandler SystemgeMessageHandler.MessageHandler) error {
	if messageHandler == nil {
		return Error.New("no message handler set", nil)
	}
	if message.GetSyncToken() == "" {
		err := messageHandler.HandleAsyncMessage(message)
		if err != nil {
			return Error.New("failed to handle async message", err)
		}
	} else {
		if responsePayload, err := messageHandler.HandleSyncRequest(message); err != nil {
			if err := connection.send(message.NewFailureResponse(err.Error()).Serialize()); err != nil {
				return Error.New("failed to send failure response", err)
			}
		} else {
			if err := connection.send(message.NewSuccessResponse(responsePayload).Serialize()); err != nil {
				return Error.New("failed to send success response", err)
			}
		}
	}
	return nil
}

func (connection *SystemgeConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
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
