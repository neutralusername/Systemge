package TcpSystemgeConnection

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

// A started loop will run until stopChannel receives a value (or is closed) or connection.GetNextMessage returns an error.
// errorChannel will send all errors that occur during message processing.
func (connection *TcpSystemgeConnection) StartMessageHandlingLoop_Sequentially(messageHandler SystemgeConnection.MessageHandler) error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.StartingMessageHandlingLoop,
		"starting message handling loop",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.MessageHandlingLoop,
			Event.Behaviour:     Event.Sequential,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
			Event.ChannelType:   Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.messageHandlingLoopStopChannel != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStarted,
			"message handling loop already started",
			Event.Context{
				Event.Circumstance:  Event.MessageHandlingLoop,
				Event.Behaviour:     Event.Sequential,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
				Event.ChannelType:   Event.MessageChannel,
			},
		))
		return errors.New("message handling loop already started")
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
			case message := <-connection.messageChannel:
				if message == nil {
					if connection.infoLogger != nil {
						connection.infoLogger.Log("Connection closed and no remaining messages")
					}
					connection.StopMessageHandlingLoop()
					return
				}
				connection.messageChannelSemaphore.ReleaseBlocking()
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
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
		return errors.New("Message handling loop already started")
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
			case message := <-connection.messageChannel:
				if message == nil {
					if connection.infoLogger != nil {
						connection.infoLogger.Log("Connection closed and no remaining messages")
					}
					connection.StopMessageHandlingLoop()
					return
				}
				connection.messageChannelSemaphore.ReleaseBlocking()
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
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
		return Event.New("no message handler set", nil)
	}
	if message.GetSyncToken() == "" {
		err := messageHandler.HandleAsyncMessage(connection, message)
		if err != nil {
			return Event.New("failed to handle async message", err)
		}
	} else {
		if message.IsResponse() {
			return Event.New("message is a response, cannot handle", nil)
		}
		if responsePayload, err := messageHandler.HandleSyncRequest(connection, message); err != nil {
			if err := connection.SyncResponse(message, false, err.Error()); err != nil {
				return Event.New("failed to send failure response", err)
			}
		} else {
			if err := connection.SyncResponse(message, true, responsePayload); err != nil {
				return Event.New("failed to send success response", err)
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
		return Event.New("Message handling loop not registered", nil)
	}
	close(connection.messageHandlingLoopStopChannel)
	connection.messageHandlingLoopStopChannel = nil
	return nil
}

func (connection *TcpSystemgeConnection) GetNextMessage() (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel != nil {
		return nil, Event.New("Message handling loop is registered", nil)
	}
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}
	select {
	case message := <-connection.messageChannel:
		if message == nil {
			return nil, Event.New("Connection closed and no remaining messages", nil)
		}
		connection.messageChannelSemaphore.ReleaseBlocking()
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
		}
		return message, nil
	case <-timeout:
		return nil, Event.New("Timeout while waiting for message", nil)
	}
}

func (connection *TcpSystemgeConnection) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}
