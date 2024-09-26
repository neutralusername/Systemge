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
func (connection *TcpSystemgeConnection) StartMessageHandlingLoop(messageHandler SystemgeConnection.MessageHandler, sequentially bool) error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	var behaviour string
	if sequentially {
		behaviour = Event.Sequential
	} else {
		behaviour = Event.Concurrent
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStarting,
		"starting message handling loop",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.StartMessageHandlingLoop,
			Event.Behaviour:     behaviour,
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
				Event.Circumstance:  Event.StartMessageHandlingLoop,
				Event.Behaviour:     behaviour,
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

	go connection.messageHandlingLoop(stopChannel, messageHandler, sequentially)

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStarted,
		"message handling loop started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.StartMessageHandlingLoop,
			Event.Behaviour:     behaviour,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
			Event.ChannelType:   Event.MessageChannel,
		},
	)); !event.IsInfo() {
		close(connection.messageHandlingLoopStopChannel)
		connection.messageHandlingLoopStopChannel = nil
		return event.GetError()
	}

	return nil
}

func (connection *TcpSystemgeConnection) messageHandlingLoop(stopChannel chan bool, messageHandler SystemgeConnection.MessageHandler, sequentially bool) {
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
			if sequentially {
				if err := connection.HandleMessage(message, messageHandler); err != nil {
					if connection.errorLogger != nil {
						connection.errorLogger.Log(err.Error())
					}
				}
			} else {
				go func() {
					if err := connection.HandleMessage(message, messageHandler); err != nil {
						if connection.errorLogger != nil {
							connection.errorLogger.Log(err.Error())
						}
					}
				}()
			}
		}
	}
}

func (connection *TcpSystemgeConnection) StopMessageHandlingLoop() error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel == nil {
		return errors.New("Message handling loop not registered")
	}
	close(connection.messageHandlingLoopStopChannel)
	connection.messageHandlingLoopStopChannel = nil
	return nil
}

func (connection *TcpSystemgeConnection) IsMessageHandlingLoopStarted() bool {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	return connection.messageHandlingLoopStopChannel != nil
}

func (connection *TcpSystemgeConnection) GetNextMessage() (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	if connection.messageHandlingLoopStopChannel != nil {
		return nil, errors.New("Message handling loop is registered")
	}
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}
	select {
	case message := <-connection.messageChannel:
		if message == nil {
			return nil, errors.New("Connection closed and no remaining messages")
		}
		connection.messageChannelSemaphore.ReleaseBlocking()
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
		}
		return message, nil
	case <-timeout:
		return nil, errors.New("Timeout while waiting for message")
	}
}

func (connection *TcpSystemgeConnection) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}

// HandleMessage will determine if the message is synchronous or asynchronous and call the appropriate handler.
func (connection *TcpSystemgeConnection) HandleMessage(message *Message.Message, messageHandler SystemgeConnection.MessageHandler) error {
	if messageHandler == nil {
		return errors.New("no message handler set")
	}
	if message.GetSyncToken() == "" {
		err := messageHandler.HandleAsyncMessage(connection, message)
		if err != nil {
			return err
		}
	} else {
		if message.IsResponse() {
			return errors.New("message is a response, cannot handle")
		}
		if responsePayload, err := messageHandler.HandleSyncRequest(connection, message); err != nil {
			if err := connection.SyncResponse(message, false, err.Error()); err != nil {
				return err
			}
		} else {
			if err := connection.SyncResponse(message, true, responsePayload); err != nil {
				return err
			}
		}
	}
	return nil
}
