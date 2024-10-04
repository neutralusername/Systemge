package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *WebsocketClient) RetrieveNextMessage(waitDurationMs uint64) (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	if connection.messageHandlingLoopStopChannel != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStarted,
			"message handling loop is registered",
			Event.Context{
				Event.Circumstance: Event.RetrieveNextMessage,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return nil, errors.New("message handling loop is registered")
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"receiving from message channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.RetrieveNextMessage,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	var timeout <-chan time.Time
	if waitDurationMs > 0 {
		timeout = time.After(time.Duration(waitDurationMs) * time.Millisecond)
	}
	select {
	case message := <-connection.messageChannel:
		if message == nil {
			connection.onEvent(Event.NewWarningNoOption(
				Event.ReceivedNilValueFromChannel,
				"received nil value from message channel",
				Event.Context{
					Event.Circumstance: Event.RetrieveNextMessage,
					Event.ChannelType:  Event.MessageChannel,
				},
			))
			return nil, errors.New("received nil value from message channel")
		}
		if event := connection.onEvent(Event.NewInfo(
			Event.ReceivedFromChannel,
			"received message from message channel",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.RetrieveNextMessage,
				Event.ChannelType:  Event.MessageChannel,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		)); !event.IsInfo() {
			return nil, event.GetError()
		}
		connection.messageChannelSemaphore.ReleaseBlocking()
		return message, nil
	case <-timeout:
		connection.onEvent(Event.NewWarningNoOption(
			Event.Timeout,
			"timeout while waiting for message",
			Event.Context{
				Event.Circumstance: Event.RetrieveNextMessage,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return nil, errors.New("timeout while waiting for message")
	}
}

/*
func (server *WebsocketServer) toTopicHandler(handler WebsocketClient.WebsocketMessageHandler) Tools.TopicHandler {
	return func(args ...any) (any, error) {
		websocketConnection := args[0].(*WebsocketClient.WebsocketClient)
		message := args[1].(*Message.Message)

		if server.eventHandler != nil {
			if event := server.onEvent(Event.NewInfo(
				Event.HandlingMessage,
				"handling websocketConnection message",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.HandleMessage,
					Event.Identity:     websocketConnection.GetName(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				return nil, errors.New("event cancelled")
			}
		}

		err := handler(websocketConnection, message)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.HandleMessageFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.HandleMessage,
						Event.Identity:     websocketConnection.GetName(),
						Event.Address:      websocketConnection.GetAddress(),
						Event.Topic:        message.GetTopic(),
						Event.Payload:      message.GetPayload(),
						Event.SyncToken:    message.GetSyncToken(),
					},
				))
			}
			return nil, err
		}

		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.HandledMessage,
				"handled websocketConnection message",
				Event.Context{
					Event.Circumstance: Event.HandleMessage,
					Event.Identity:     websocketConnection.GetName(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
		}

		return nil, nil
	}
}

		topicHandlers := make(Tools.TopicHandlers)
	   	for topic, handler := range messageHandlers {
	   		topicHandlers[topic] = server.toTopicHandler(handler)
	   	}
	   	server.topicManager = Tools.NewTopicManager(config.TopicManagerConfig, topicHandlers, nil) */

// A started loop will run until stopChannel receives a value (or is closed) or connection.GetNextMessage returns an error.
// errorChannel will send all errors that occur during message processing.
func (connection *WebsocketClient) StartMessageHandlingLoop(topicManager *Tools.TopicManager, sequentially bool) error {
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
			Event.Circumstance: Event.MessageHandlingLoopStart,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.messageHandlingLoopStopChannel != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStarted,
			"message handling loop already started",
			Event.Context{
				Event.Circumstance: Event.MessageHandlingLoopStart,
				Event.Behaviour:    behaviour,
				Event.ChannelType:  Event.MessageChannel,
			},
		))
		return errors.New("message handling loop already started")
	}
	stopChannel := make(chan bool)
	connection.messageHandlingLoopStopChannel = stopChannel

	go connection.messageHandlingLoop(stopChannel, topicManager, sequentially, behaviour)

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStarted,
		"message handling loop started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoopStart,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		close(connection.messageHandlingLoopStopChannel)
		connection.messageHandlingLoopStopChannel = nil
		return event.GetError()
	}

	return nil
}

func (connection *WebsocketClient) messageHandlingLoop(stopChannel chan bool, topicManager *Tools.TopicManager, sequentially bool, behaviour string) {
	if event := connection.onEvent(Event.NewInfo(
		Event.ReceivingFromChannel,
		"message handling loop running",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoop,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
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
						Event.Circumstance: Event.MessageHandlingLoop,
						Event.ChannelType:  Event.MessageChannel,
						Event.Behaviour:    behaviour,
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
					Event.Circumstance: Event.MessageHandlingLoop,
					Event.Behaviour:    behaviour,
					Event.ChannelType:  Event.MessageChannel,
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				connection.StopMessageHandlingLoop()
				return
			}

			if sequentially {
				topicManager.HandleTopic(message.GetTopic(), message, connection)
			} else {
				go func() {
					topicManager.HandleTopic(message.GetTopic(), message, connection)
				}()
			}
		}
	}
}

func (connection *WebsocketClient) StopMessageHandlingLoop() error {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageHandlingLoopStopping,
		"stopping message handling loop",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageHandlingLoopStop,
			Event.ChannelType:  Event.MessageChannel,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if connection.messageHandlingLoopStopChannel == nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.MessageHandlingLoopAlreadyStopped,
			"message handling loop already stopped",
			Event.Context{
				Event.Circumstance: Event.MessageHandlingLoopStop,
				Event.ChannelType:  Event.MessageChannel,
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
			Event.Circumstance: Event.MessageHandlingLoopStop,
			Event.ChannelType:  Event.MessageChannel,
		},
	))
	return nil
}

func (connection *WebsocketClient) IsMessageHandlingLoopStarted() bool {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	return connection.messageHandlingLoopStopChannel != nil
}

func (connection *WebsocketClient) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}
