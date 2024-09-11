package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *TcpConnection) UnprocessedMessagesCount() int64 {
	return connection.unprocessedMessages.Load()
}

func (connection *TcpConnection) GetNextMessage() (*Message.Message, error) {
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
		if message == nil {
			return nil, Error.New("Connection closed and no remaining messages", nil)
		}
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Returned message # in GetNextMessage" + Helpers.Uint64ToString(message.id))
		}
		connection.unprocessedMessages.Add(-1)
		return message.message, nil
	case <-timeout:
		return nil, Error.New("Timeout while waiting for message", nil)
	}
}

func (connection *TcpConnection) StopProcessingLoop() error {
	connection.processMutex.Lock()
	defer connection.processMutex.Unlock()
	if connection.processingLoopStopChannel == nil {
		return Error.New("Processing loop not running", nil)
	}
	close(connection.processingLoopStopChannel)
	connection.processingLoopStopChannel = nil
	return nil
}

// A started loop will run indefinitely until StopProcessingLoop is called.
func (connection *TcpConnection) StartProcessingLoopSequentially(messageHandler SystemgeConnection.MessageHandler) error {
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
				if message == nil {
					if connection.infoLogger != nil {
						connection.infoLogger.Log("Reached end of messages. Stopping processing messages sequentially")
					}
					connection.StopProcessingLoop()
					return
				}
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

// A started loop will run indefinitely until StopProcessingLoop is called.
func (connection *TcpConnection) StartProcessingLoopConcurrently(messageHandler SystemgeConnection.MessageHandler) error {
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
				if message == nil {
					if connection.infoLogger != nil {
						connection.infoLogger.Log("Reached end of messages. Stopping processing messages concurrently")
					}
					connection.StopProcessingLoop()
					return
				}
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

func (connection *TcpConnection) ProcessMessage(message *Message.Message, messageHandler SystemgeConnection.MessageHandler) error {
	if messageHandler == nil {
		return Error.New("no message handler set", nil)
	}
	if message.GetSyncToken() == "" {
		err := messageHandler.HandleAsyncMessage(connection, message)
		if err != nil {
			return Error.New("failed to handle async message", err)
		}
	} else {
		if responsePayload, err := messageHandler.HandleSyncRequest(connection, message); err != nil {
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
