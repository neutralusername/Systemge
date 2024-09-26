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
			Event.Circumstance:  Event.MessageHandlingLoopStart,
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
				Event.Circumstance:  Event.MessageHandlingLoopStart,
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

	go connection.messageHandlingLoop(stopChannel, messageHandler, sequentially, behaviour)

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStarted,
		"message handling loop started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.MessageHandlingLoopStart,
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

func (connection *TcpSystemgeConnection) messageHandlingLoop(stopChannel chan bool, messageHandler SystemgeConnection.MessageHandler, sequentially bool, behaviour string) {
	if event := connection.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"message handling loop running",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.MessageHandlingLoop,
			Event.Behaviour:     behaviour,
			Event.ChannelType:   Event.MessageChannel,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	)); !event.IsInfo() {
		connection.StopMessageHandlingLoop()
		return
	}

	for {
		select {
		case <-stopChannel:
			return
		case message := <-connection.messageChannel:

			if message == nil {
				connection.onEvent(Event.NewInfoNoOption(
					Event.ReceivedNilValueFromChannel,
					"received nil value from message channel",
					Event.Context{
						Event.Circumstance:  Event.MessageHandlingLoop,
						Event.ChannelType:   Event.MessageChannel,
						Event.Behaviour:     behaviour,
						Event.ClientType:    Event.TcpSystemgeConnection,
						Event.ClientName:    connection.GetName(),
						Event.ClientAddress: connection.GetAddress(),
					},
				))
				connection.StopMessageHandlingLoop()
				return
			}

			connection.messageChannelSemaphore.ReleaseBlocking()
			if event := connection.onEvent(Event.NewInfo(
				Event.ReceivedFromChannel,
				"received message from message channel",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.MessageHandlingLoop,
					Event.Behaviour:     behaviour,
					Event.ChannelType:   Event.MessageChannel,
					Event.ClientType:    Event.TcpSystemgeConnection,
					Event.ClientName:    connection.GetName(),
					Event.ClientAddress: connection.GetAddress(),
					Event.Topic:         message.GetTopic(),
					Event.Payload:       message.GetPayload(),
					Event.SyncToken:     message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				connection.StopMessageHandlingLoop()
				return
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

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStopping,
		"stopping message handling loop",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.MessageHandlingLoopStop,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
			Event.ChannelType:   Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.messageHandlingLoopStopChannel == nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStopped,
			"message handling loop already stopped",
			Event.Context{
				Event.Circumstance:  Event.MessageHandlingLoopStop,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
				Event.ChannelType:   Event.MessageChannel,
			},
		))
		return errors.New("message handling loop already stopped")
	}

	close(connection.messageHandlingLoopStopChannel)
	connection.messageHandlingLoopStopChannel = nil

	connection.onEvent(Event.NewInfoNoOption(
		Event.MessageHandlingLoopStopped,
		"message handling loop stopped",
		Event.Context{
			Event.Circumstance:  Event.MessageHandlingLoopStop,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
			Event.ChannelType:   Event.MessageChannel,
		},
	))
	return nil
}

func (connection *TcpSystemgeConnection) IsMessageHandlingLoopStarted() bool {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	return connection.messageHandlingLoopStopChannel != nil
}

func (connection *TcpSystemgeConnection) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}

func (connection *TcpSystemgeConnection) RetrieveNextMessage() (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	if connection.messageHandlingLoopStopChannel != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStarted,
			"message handling loop is registered",
			Event.Context{
				Event.Circumstance:  Event.RetrievingNextMessage,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
				Event.ChannelType:   Event.MessageChannel,
			},
		))
		return nil, errors.New("Message handling loop is registered")
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"receiving from message channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.RetrievingNextMessage,
			Event.ChannelType:   Event.MessageChannel,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
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
