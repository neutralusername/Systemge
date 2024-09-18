package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

// A started loop will run until stopChannel receives a value (or is closed) or connection.GetNextMessage returns an error.
// errorChannel will send all errors that occur during message processing.
func (connection *TcpSystemgeConnection) StartMessageHandlingLoop_Sequentially(messageHandler SystemgeConnection.MessageHandler) error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel != nil {
		return Error.New("Message handling loop already registered", nil)
	}
	stopChann := make(chan bool)
	connection.messageHandlingLoopStopChannel = stopChann
	go func() {
		for {
			select {
			case <-stopChann:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Message handling loop stopped")
				}
				return
			default:
				message, err := connection.GetNextMessage()
				if err != nil {
					if connection.warningLogger != nil {
						connection.warningLogger.Log(err.Error())
					}
					connection.StopMessageHandlingLoop()
					return
				}
				if err := connection.HandleMessage(message, messageHandler); err != nil {
					if connection.errorLogger != nil {
						connection.errorLogger.Log(err.Error())
					}
				}
			}
		}
	}()
	return nil
}

// A started loop will run until stopChannel receives a value (or is closed) or connection.GetNextMessage returns an error.
// errorChannel will send all errors that occur during message processing.
func (connection *TcpSystemgeConnection) StartMessageHandlingLoop_Concurrently(messageHandler SystemgeConnection.MessageHandler) error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel != nil {
		return Error.New("Message handling loop already registered", nil)
	}
	stopChannel := make(chan bool)
	connection.messageHandlingLoopStopChannel = stopChannel
	go func() {
		for {
			select {
			case <-stopChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Message handling loop stopped")
				}
				return
			default:
				message, err := connection.GetNextMessage()
				if err != nil {
					if connection.warningLogger != nil {
						connection.warningLogger.Log(err.Error())
					}
					connection.StopMessageHandlingLoop()
					return
				}
				go func() {
					if err := connection.HandleMessage(message, messageHandler); err != nil {
						if connection.errorLogger != nil {
							connection.errorLogger.Log(err.Error())
						}
					}
				}()
			}
		}
	}()
	return nil
}

// HandleMessage will determine if the message is synchronous or asynchronous and call the appropriate handler.
func (connection *TcpSystemgeConnection) HandleMessage(message *Message.Message, messageHandler SystemgeConnection.MessageHandler) error {
	if messageHandler == nil {
		return Error.New("no message handler set", nil)
	}
	if message.GetSyncToken() == "" {
		err := messageHandler.HandleAsyncMessage(connection, message)
		if err != nil {
			return Error.New("failed to handle async message", err)
		}
	} else {
		if message.IsResponse() {
			return Error.New("message is a response, cannot handle", nil)
		}
		if responsePayload, err := messageHandler.HandleSyncRequest(connection, message); err != nil {
			if err := connection.SyncResponse(message, false, err.Error()); err != nil {
				return Error.New("failed to send failure response", err)
			}
		} else {
			if err := connection.SyncResponse(message, true, responsePayload); err != nil {
				return Error.New("failed to send success response", err)
			}
		}
	}
	return nil
}

func (connection *TcpSystemgeConnection) IsMessageHandlingLoopStarted() bool {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	return connection.messageHandlingLoopStopChannel != nil
}

func (connection *TcpSystemgeConnection) StopMessageHandlingLoop() error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel == nil {
		return Error.New("Message handling loop not registered", nil)
	}
	close(connection.messageHandlingLoopStopChannel)
	connection.messageHandlingLoopStopChannel = nil
	return nil
}

func (connection *TcpSystemgeConnection) GetNextMessage() (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel != nil {
		return nil, Error.New("Message handling loop is registered", nil)
	}
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}

	select {
	case message := <-connection.messageChannel:
		if message == nil {
			return nil, Error.New("Connection closed and no remaining messages", nil)
		}
		connection.messageChannelSemaphore.ReleaseBlocking()
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
		}
		return message, nil
	case <-timeout:
		return nil, Error.New("Timeout while waiting for message", nil)
	}
}

func (connection *TcpSystemgeConnection) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}

func (connection *TcpSystemgeConnection) addMessageToChannel(messageBytes []byte) error {
	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		connection.byteRateLimiterExceeded.Add(1)
		return Error.New("byte rate limiter exceeded", nil)
	}
	if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
		connection.messageRateLimiterExceeded.Add(1)
		return Error.New("message rate limiter exceeded", nil)
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
			connection.invalidSyncResponsesReceived.Add(1)
			return Error.New("failed to add sync response", err)
		}
		connection.validMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return nil
	} else {
		connection.validMessagesReceived.Add(1)
		connection.messageChannel <- message
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Added message \"" + Helpers.GetPointerId(message) + "\" to processing channel")
		}
		return nil
	}
}
